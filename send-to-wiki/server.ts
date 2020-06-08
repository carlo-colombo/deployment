import { log, listenAndServe, decoder } from "./deps.ts"

const tagRe = /(?:^|[^\S\xA0])(?:\[\[(.*?)\]\])(?=[^\S\xA0]|$)|([\S\xA0]+)/gm

const ALLOWED_CHATS = Deno.env.get("ALLOWED_CHATS") || ""
const WIKI_URL = Deno.env.get("WIKI_URL")
const USERNAME = Deno.env.get("WIKI_USERNAME")
const PASSWORD = Deno.env.get("WIKI_PASSWORD")
const PORT = Deno.env.get("PORT") || 8080
const EXTRACT_INFO_URL = Deno.env.get("EXTRACT_INFO_URL") || ""
const TELEGRAM_TOKEN = Deno.env.get("TELEGRAM_TOKEN")

const address = `0.0.0.0:${PORT}`

log.info(`Listening ${address} ...`)

interface Update {
  message: {
    chat: {
      id: number
    },
    text: string,
    entities: Array<{
      offset: number,
      length: number,
      type: "url"
    }>
  }
}


const headers = {
  'X-Requested-With': 'TiddlyWiki',
  'authorization': `Basic ${btoa(`${USERNAME}:${PASSWORD}`)}`
}


function wiki(method: 'PUT' | 'GET', title: string, data?: object | null): Promise<Response> {
  return fetch(`${WIKI_URL}/recipes/default/tiddlers/${encodeURIComponent(title)}`, {
    method,
    headers,
    body: JSON.stringify(data)
  })
}

let title = ""

listenAndServe(address, async (req) => {
  try {
    const data = decoder.decode(await Deno.readAll(req.body));

    const {
      message: {
        chat: { id },
        text,
        entities
      }
    }: Update = JSON.parse(data);

    if (ALLOWED_CHATS != "" && !ALLOWED_CHATS.includes(id.toString())) {
      log.error(`Invalid chat (${id}): ${text}`)
      req.respond({ status: 201 })
      return
    }

    const urlEntity = entities && entities.find(e => e.type == "url")


    if (urlEntity) {
      const url = text.substr(urlEntity.offset, urlEntity.length)

      console.log(EXTRACT_INFO_URL)

      const info = await fetch(EXTRACT_INFO_URL, {
        method: 'POST',
        body: JSON.stringify({ url }),
        headers: {
          'content-type': 'application/json'
        }
      }).then(resp => resp.json())

      title = info.title
      log.info(`Collected information for url: ${url}, title: ${title}`)

      const { status } = await wiki('PUT', title, { text: text, tags: 'external' })

      log.info(`Creating: '${title}', response: ${status}`)

      const { status: statusSend, statusText} = await fetch(`https://api.telegram.org/bot${TELEGRAM_TOKEN}/sendMessage`, {
        method: 'POST',
        body: JSON.stringify({
          chat_id: id,
          text: `Entry posted: ${WIKI_URL}/#${encodeURIComponent(title)}`
        }),
        headers: {
          'content-type': 'application/json'
        }
      })

      log.info(`Answering: '${title}', response: ${statusSend}, ${statusText}`)
    } else if (title != "") {
      const tiddly: {
        tags: string, title: string
      } = await wiki('GET', title).then(resp => resp.json())

      log.info(`Adding tags '${text}' to ${title}`)
      title = ""

      const { status } = await wiki('PUT', tiddly.title, {
        ...tiddly,
        tags: tiddly.tags + " " + text
      })

      log.info(`Response: ${status}`)
    } else {
      log.info('Neither url or title to attach tags')
    }
  }
  catch (e) {
    console.log(e)
    log.error(e)
  }
  finally {
    req.respond({ status: 201 })
  }
})