import {log} from "./deps.ts" 

try {
    const BOT_URL = Deno.env.get('BOT_URL')
    const TELEGRAM_TOKEN = Deno.env.get('TELEGRAM_TOKEN')

    const { status } = await fetch(`https://api.telegram.org/bot${TELEGRAM_TOKEN}/setWebhook`, {
        method: 'POST',
        headers: {
            'Content-type': 'application/json'
        },
        body: JSON.stringify({
            url: BOT_URL
        })
    })

    const resp = await fetch(`https://api.telegram.org/bot${TELEGRAM_TOKEN}/getWebhookInfo`)
    log.info(`setWebhook (${status}): ${BOT_URL}`)
    log.info(JSON.stringify(await resp.json()))
} catch (e) {
    log.error(`Error setting webhook: ${e}`)
}
