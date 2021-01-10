defmodule TiddlyWiki.Client do
  @behaviour TiddlyWiki

  alias TiddlyWiki.Tiddler

  defmacrop accept_status_code(accepted) do
    quote do
      case var!(status_code) do
        unquote(accepted) -> {:ok, var!(body)}
        _ -> {:error, %{status_code: var!(status_code)}}
      end
    end
  end

  @impl TiddlyWiki
  def get_all(
        %TiddlyWiki{
          url: url,
          username: username,
          password: password
        },
        filter \\ ""
      ) do
    with {:ok, resp} <-
           (url <> "/recipes/default/tiddlers.json")
           |> URI.parse()
           |> Map.put(:query, URI.encode_query(%{"filter" => filter}))
           |> URI.to_string()
           |> Mojito.get(headers(username, password)),
         %{status_code: status_code, body: body} <- resp,
         {:ok, body} <- accept_status_code(200),
         {:ok, tiddlers} <- Jason.decode(body, keys: :atoms) do
      {:ok,
       tiddlers
       |> Enum.map(&struct(Tiddler, &1))}
    end
  end

  @impl TiddlyWiki
  def put(
        %TiddlyWiki{
          url: url,
          username: username,
          password: password,
          external_url: external_url
        },
        %Tiddler{title: title} = tiddler
      ) do
    with {:ok, resp} <-
           url
           |> URI.merge("recipes/default/tiddlers/" <> encode(title))
           |> URI.to_string()
           |> Mojito.put(
             headers(
               username,
               password,
               [{"content-type", "application/json"}]
             ),
             Jason.encode!(%{tiddler | creator: username, modifier: username})
           ),
         %{status_code: status_code, body: body} <- resp,
         {:ok, _} <- accept_status_code(204) do
      {:ok,
       external_url
       |> URI.merge("#" <> encode(title))
       |> URI.to_string()}
    end
  end

  @impl TiddlyWiki
  def get(
        %TiddlyWiki{
          url: url,
          username: username,
          password: password
        },
        title
      ) do
    with {:ok, resp} <-
           url
           |> URI.merge("recipes/default/tiddlers/" <> encode(title))
           |> URI.to_string()
           |> Mojito.get(headers(username, password)),
         %{status_code: status_code, body: body} <- resp,
         {:ok, body} <- accept_status_code(200),
         {:ok, tiddler} <- Jason.decode(body, keys: :atoms) do
      {:ok, struct(Tiddler, tiddler)}
    end
  end

  defp headers(username, password, additional \\ []) do
    [
      Mojito.Headers.auth_header(username, password),
      {"X-Requested-With", "TiddlyWiki"}
    ] ++ additional
  end

  defp encode(s) do
    URI.encode(s, fn
      ?: -> false
      ?/ -> false
      c -> URI.char_unescaped?(c)
    end)
  end
end
