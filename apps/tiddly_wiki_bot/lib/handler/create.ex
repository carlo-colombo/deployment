defmodule TiddlyWikiBot.Handler.Create do
  import TiddlyWikiBot.Handler,
    only: [
      get_tg_client: 0,
      get_tw_client: 0,
      new_tiddler: 3,
      escape_markdown_v2: 1
    ]

  def init(opts), do: opts

  def call(conn, _opts) do
    chat_id = conn.assigns[:chat_id]
    "/create " <> text = conn.assigns[:text]

    with tiddler <- new_tiddler(text, text, chat_id),
         {:ok, url} <- TiddlyWiki.put(tiddler),
         {:ok, _} <-
           chat_id
           |> get_tg_client().send_message(
             escape_markdown_v2(url),
             parse_mode: "MarkdownV2"
           ) do
      url
    else
      e -> {:error, e}
    end

    conn
  end
end
