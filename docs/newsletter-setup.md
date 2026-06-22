# Newsletter Ingestion Setup (Cloudflare)

Receive newsletters as articles in blogwatcher using Cloudflare Email Routing, a Cloudflare Worker,
and a Cloudflare Tunnel — all on the free tier.

## How it works

```
Newsletter sender → Cloudflare Email Routing → Email Worker
                                                     ↓ HTTP POST (raw RFC 822)
                              blogwatcher ← Cloudflare Tunnel ← /newsletter/webhook
```

Each new sender address auto-creates a blog entry; subsequent emails from the same sender
appear as articles under that blog.

---

## Prerequisites

- A Cloudflare account (free tier is sufficient)
- A domain whose DNS is managed by Cloudflare
  - If your root domain already uses Google Workspace or another mail provider,
    use a **subdomain** for Email Routing (e.g. `mail.yourdomain.com`) so MX records
    for the root are not touched
- [Node.js](https://nodejs.org/) and `npx` for running `wrangler`
- [`cloudflared`](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/)
  installed on the host machine running blogwatcher

---

## Step 1 — Enable Cloudflare Email Routing

1. Log in to the [Cloudflare dashboard](https://dash.cloudflare.com) and select your domain
   (or subdomain zone).
2. Go to **Email → Email Routing** and click **Enable Email Routing**.
3. Add a **catch-all rule**:
   - **Action**: Send to a Worker
   - **Destination**: the worker you will create in Step 2 (name it `blogwatcher-email`)
4. Save. Cloudflare will add the necessary MX records automatically.

> **Google Workspace note**: if your root domain already has Google Workspace MX records,
> create a separate Cloudflare zone for a subdomain (e.g. `mail.yourdomain.com`) and
> configure Email Routing there. Use `yourname@mail.yourdomain.com` as your newsletter
> inbox address.

---

## Step 2 — Create the Email Worker

### `worker.js`

```js
export default {
  async email(message, env) {
    const raw = await new Response(message.raw).arrayBuffer();
    const body = new Uint8Array(raw);

    const res = await fetch(env.WEBHOOK_URL, {
      method: "POST",
      headers: {
        "Content-Type": "message/rfc822",
        "X-Webhook-Secret": env.WEBHOOK_SECRET,
      },
      body,
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

[vars]
# Set WEBHOOK_URL and WEBHOOK_SECRET as secrets (see below)
```

### Deploy

```bash
# Store secrets — never commit these to version control
npx wrangler secret put WEBHOOK_URL
# When prompted, enter: https://bw.yourdomain.com/newsletter/webhook

npx wrangler secret put WEBHOOK_SECRET
# When prompted, paste the secret from blogwatcher Settings → Newsletter Inbox

npx wrangler deploy
```

---

## Step 3 — Set up Cloudflare Tunnel

### Create the tunnel

```bash
cloudflared tunnel create blogwatcher
```

This writes a credentials file to `~/.cloudflared/<tunnel-id>.json`.

### Tunnel configuration

Create `~/.cloudflared/config.yml`:

```yaml
tunnel: blogwatcher
credentials-file: /home/<user>/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: bw.yourdomain.com
    service: http://localhost:8080
  - service: http_status:404
```

Replace `<user>`, `<tunnel-id>`, and `8080` with your actual values.

### Route DNS

```bash
cloudflared tunnel route dns blogwatcher bw.yourdomain.com
```

### Run as a systemd service (Linux / Raspberry Pi)

```bash
sudo cloudflared service install
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
```

Verify: `sudo systemctl status cloudflared`

---

## Step 4 — Configure blogwatcher

1. Open **Settings** in blogwatcher.
2. Scroll to the **Newsletter Inbox** section.
3. Copy the **Webhook URL** and **Webhook Secret** — use these as the values for
   `WEBHOOK_URL` and `WEBHOOK_SECRET` when running `wrangler secret put` in Step 2.
4. Enter your newsletter inbox address (e.g. `yourname@mail.yourdomain.com`) in the
   **Inbox Email Address** field and click **Save**.

---

## Step 5 — Subscribe to your first newsletter

Use your inbox address when signing up to any newsletter service. When the first email
arrives:

- blogwatcher automatically creates a new blog entry for that sender.
- The email body appears as a readable article in your feed, alongside your RSS subscriptions.
- Subsequent emails from the same sender are added to the same blog.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| Worker throws on deploy | `WEBHOOK_URL` / `WEBHOOK_SECRET` not set | Run `wrangler secret put` for both |
| Webhook returns 401 | Secret mismatch | Re-copy the secret from Settings and re-run `wrangler secret put WEBHOOK_SECRET` |
| Tunnel not reachable | `cloudflared` not running | `sudo systemctl start cloudflared` |
| Emails not arriving | Email Routing catch-all not saved | Re-check the catch-all rule in Cloudflare dashboard |
