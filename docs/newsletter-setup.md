# Newsletter Ingestion Setup (Cloudflare)

Receive newsletters as articles in blogwatcher using Cloudflare Email Routing and a Cloudflare Email Worker — both free tier.

## How it works

```
Newsletter sender
  → Cloudflare Email Routing (MX on subdomain)
  → Email Worker (forwards raw email via HTTP POST)
  → Cloudflare Tunnel (exposes ONLY /newsletter/webhook, nothing else)
  → blogwatcher POST /newsletter/webhook
  → stored as article in your feed
```

The tunnel is locked down two ways:
1. **Path restriction** — cloudflared itself returns 404 for every path except `/newsletter/webhook`. The blogwatcher UI is unreachable from the internet.
2. **Webhook secret** — blogwatcher rejects any POST that doesn't include the correct `X-Webhook-Secret` header. The secret is a 32-byte random value generated at first run.

---

## Prerequisites

- Cloudflare account (free tier) with your domain's DNS managed by Cloudflare
- [`cloudflared`](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) installed on the machine running blogwatcher
- [Node.js](https://nodejs.org/) with `npx` (for `wrangler`)
- blogwatcher running locally

---

## Step 1 — Create the tunnel (do not start it yet)

```bash
cloudflared tunnel create blogwatcher
```

Credentials are written to `~/.cloudflared/<tunnel-id>.json`. Note the tunnel ID.

### `~/.cloudflared/config.yml`

```yaml
tunnel: blogwatcher
credentials-file: /home/<user>/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: newsletters.yourdomain.com
    path: /newsletter/webhook    # <-- THIS LINE is what blocks the rest of the server
    service: http://localhost:8080
  - service: http_status:404
```

Replace `<user>`, `<tunnel-id>`, `newsletters.yourdomain.com`, and `8080` with your values.

The `path:` field is critical — without it, cloudflared routes ALL traffic on that hostname to blogwatcher and your entire RSS server is public. With it, only `/newsletter/webhook` is routed; everything else returns 404 from cloudflared itself.

After editing the file, restart cloudflared:

```bash
sudo systemctl restart cloudflared
```

Verify it worked — this should return 404, not your blogwatcher UI:

```bash
curl -I https://newsletters.yourdomain.com/
```

### Route DNS

```bash
cloudflared tunnel route dns blogwatcher newsletters.yourdomain.com
```

**Do not start the tunnel yet.** Set up the service token in Step 2 first.

---

## Step 2 — Start the tunnel

```bash
sudo cloudflared service install
sudo systemctl enable --now cloudflared
```

Your webhook URL is: **`https://newsletters.yourdomain.com/newsletter/webhook`**

Verify the path restriction works: `curl https://newsletters.yourdomain.com` should return 404. `curl https://newsletters.yourdomain.com/newsletter/webhook` should return 401 (no service token presented).

---

## Step 3 — Get your webhook secret from blogwatcher

1. Open blogwatcher locally and go to **Settings**.
2. Scroll to the **Newsletter Inbox** section.
3. Copy the **Webhook Secret**.
4. Set your **Inbox Email Address** (e.g. `newsletters@mail.yourdomain.com`) and click **Save**.

---

## Step 4 — Deploy the Email Worker

The worker runs on **Cloudflare's edge infrastructure**, not on your machine. You run the commands below from any machine with Node.js (your laptop is fine) — `wrangler deploy` uploads the worker to Cloudflare. After deployment it lives entirely in Cloudflare and you can delete the local files.

Now you have everything needed: the public URL (Step 2) and the webhook secret (Step 3).

### `worker.js`

```js
export default {
  async email(message, env) {
    const raw = await new Response(message.raw).arrayBuffer();

    const res = await fetch(env.WEBHOOK_URL, {
      method: "POST",
      redirect: "error",  // fail visibly if Cloudflare Access redirects instead of authenticating
      headers: {
        "Content-Type": "message/rfc822",
        "X-Webhook-Secret": env.WEBHOOK_SECRET,
        "CF-Access-Client-Id": env.CF_ACCESS_CLIENT_ID,
        "CF-Access-Client-Secret": env.CF_ACCESS_CLIENT_SECRET,
      },
      body: new Uint8Array(raw),
    });

    if (!res.ok) {
      throw new Error(`blogwatcher webhook returned ${res.status}`);
    }
  },
};
```

### `wrangler.toml`

```toml
name = "blogwatcher-email"
main = "worker.js"
compatibility_date = "2024-01-01"
```

### Set secrets and deploy

```bash
npx wrangler secret put WEBHOOK_URL
# Enter: https://newsletters.yourdomain.com/newsletter/webhook

npx wrangler secret put WEBHOOK_SECRET
# Enter: the secret copied from blogwatcher Settings

npx wrangler secret put CF_ACCESS_CLIENT_ID
# Enter: Client ID from the blogwatcher-email-worker service token

npx wrangler secret put CF_ACCESS_CLIENT_SECRET
# Enter: Client Secret from the blogwatcher-email-worker service token

npx wrangler deploy
```

---

## Step 5 — Enable Email Routing on a subdomain

> **If you have Google Workspace (or any other mail provider) on your root domain: do NOT click "Onboard Domain".** That button replaces your root domain's MX records and will break your existing email. Use a subdomain instead — the steps below add MX records only on the subdomain.

1. Go to [dash.cloudflare.com](https://dash.cloudflare.com) → **Compute → Email Service → Email Routing**.
2. Click your apex domain — do **not** click **Onboard Domain**.
3. Open the **Settings** tab.
4. Under **Subdomains**, enter `mail` (or whatever subdomain you want) and submit.
5. Cloudflare adds MX records to `mail.yourdomain.com` only. Your root MX records are untouched.

Your inbox address will be anything `@mail.yourdomain.com`.

> If the Settings tab is inaccessible without first onboarding the root domain, use a completely separate domain that has no existing email. Do not onboard a domain with active Google Workspace MX records.

---

## Step 6 — Create the Email Routing rule

1. In **Compute → Email Service → Email Routing**, select `mail.yourdomain.com`.
2. Open the **Routing Rules** tab.
3. Click **Create routing rule**.
4. Set pattern to **Catch-all**, action to **Send to a Worker**, worker to `blogwatcher-email`.
5. Click **Save**.

---

## Step 7 — Subscribe

Use your inbox address (e.g. `newsletters@mail.yourdomain.com`) when signing up to any newsletter. The first email from a new sender auto-creates a blog entry for that sender; subsequent emails land there as articles alongside your RSS feed.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| `curl newsletters.yourdomain.com` returns anything other than 404 | Tunnel ingress missing `path:` | Add `path: /newsletter/webhook` to the ingress rule in `config.yml` and restart cloudflared |
| Webhook returns 401 | Secret mismatch | Re-copy from blogwatcher Settings and re-run `wrangler secret put WEBHOOK_SECRET`, then redeploy |
| Webhook unreachable | Tunnel not running | `sudo systemctl start cloudflared` — verify with `curl https://newsletters.yourdomain.com/newsletter/webhook` |
| Emails not arriving | Routing rule missing | Re-check catch-all rule in Email Routing → Routing Rules tab |
| Root domain MX broken | "Onboard Domain" was clicked | Restore your Google Workspace MX records in Cloudflare DNS |
