defmodule TiddlyWikiBot.Handler.Upload do
  import TiddlyWikiBot.Handler,
    only: [
      get_tg_client: 0,
      get_tw_client: 0,
      new_tiddler: 3,
      escape_markdown_v2: 1
    ]

  defmacrop accept_status_code(accepted) do
    quote do
      case var!(status_code) do
        unquote(accepted) -> {:ok, var!(body)}
        _ -> {:error, %{status_code: var!(status_code)}}
      end
    end
  end

  def init(opts), do: opts

  def call(%{body_params: %{"message" => %{"document" => document}}} = conn, _opts) do
    do_work(conn, document)
  end

  def call(%{body_params: %{"message" => %{"photo" => photos}}} = conn, _opts) do
    do_work(conn, List.last(photos))
  end

  defp do_work(conn, content) do
    tg = get_tg_client()

    %{
      "file_id" => file_id,
      "file_name" => file_name
    } = content

    with {:ok, file} <- tg.get_file(file_id),
         {:ok, url} <- tg.get_file_link(file),
         {:ok, resp} <- Mojito.get(url),
         %{status_code: status_code, body: body} <- resp,
         {:ok, body} <- accept_status_code(200),
         {:ok, tiddler_url} <-
           TiddlyWiki.put(%TiddlyWiki.Tiddler{
             tags: "external",
             type: "application/" <> check_magic_bytes(body),
             title: file_name,
             text: Base.encode64(body),
             fields: %{
               chat_id: conn.assigns[:chat_id]
             }
           }),
         {:ok, _} <-
           tg.send_message(conn.assigns[:chat_id], escape_markdown_v2(tiddler_url), []) do
      conn
    end
  end

  defp check_magic_bytes(<<137, 80, 78, 71, 13, 10, 26, 10, _::binary>>), do: "png"
  defp check_magic_bytes(<<37, 80, 68, 70, 45, _::binary>>), do: "pdf"
  defp check_magic_bytes(<<73, 73, 42, 0, _::binary>>), do: "tiff"
  defp check_magic_bytes(<<77, 77, 0, 42, _::binary>>), do: "tiff"
  defp check_magic_bytes(<<255, 216, 255, _::binary>>), do: "jpg"
  defp check_magic_bytes(_), do: "unknown"
end
