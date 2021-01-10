defmodule TiddlyWikiBot.Handler.Link do
  import TiddlyWikiBot.Handler,
    only: [
      get: 1,
      insert: 2,
      get_tg_client: 0,
      get_tw_client: 0,
      get_ei_client: 0,
      new_tiddler: 3,
      escape_markdown_v2: 1
    ]

  def init(opts), do: opts

  def call(conn, _opts) do
    chat_id = conn.assigns[:chat_id]
    text = conn.assigns[:text]

    default(chat_id, text, conn.body_params)

    conn
  end

  defp default(chat_id, text, %{"message" => %{"entities" => []}}) do
    chat_id
    |> get
    |> Enum.map(&%{&1 | tags: "#{&1.tags} #{String.downcase(text)}"})
    |> Enum.map(
      &(TiddlyWiki.put(&1)
        |> (fn {:ok, url} -> url end).())
    )
  end

  defp default(chat_id, text, %{"message" => %{"entities" => entities}}) do
    entities
    |> Enum.flat_map(extract_url(text))
    |> Enum.map(&do_work(chat_id, &1))
    |> Enum.map(fn {:ok, title} -> title end)
  end

  defp default(chat_id, text, _) do
    default(chat_id, text, %{"message" => %{"entities" => []}})
  end

  defp do_work(chat_id, text) do
    with {:ok, title} <-
           (case get_ei_client().extract(text) do
              {:ok, %{title: title}} -> {:ok, title}
              {:error, _} -> {:ok, text}
            end),
         tiddler <- new_tiddler(title, text, chat_id),
         {:ok, url} <- TiddlyWiki.put(tiddler),
         :ok <- insert(chat_id, tiddler),
         {:ok, _} <-
           chat_id
           |> get_tg_client().send_message(
             escape_markdown_v2(url),
             parse_mode: "MarkdownV2"
           ) do
      {:ok, url}
    else
      e -> {:error, e}
    end
  end

  defp extract_url(text) do
    fn
      %{"type" => "url"} = entity ->
        [String.slice(text, entity["offset"], entity["length"])]

      _ ->
        []
    end
  end
end
