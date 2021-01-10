defmodule TiddlyWikiBot.Handler.Tasks do
  import TiddlyWikiBot.Handler,
    only: [
      get_tg_client: 0
    ]

  def init(opts), do: opts

  @tasks_filter Application.compile_env(:tiddlywiki_bot, :tasks_filter)

  def call(conn, _opts) do
    chat_id = conn.assigns[:chat_id]
    {:ok, res} = TiddlyWiki.get_all(@tasks_filter)

    message =
      res
      |> Enum.map(&to_item/1)
      |> Enum.join("\n")
      |> String.replace(".", "\\.")
      |> String.replace("-", "\\-")

    {:ok, _} = get_tg_client().send_message(chat_id, message, parse_mode: "MarkdownV2")

    conn
  end

  defp to_item(tiddler) do
    "[#{tiddler.title}](#{TiddlyWiki.absolute_url(tiddler)})"
  end
end
