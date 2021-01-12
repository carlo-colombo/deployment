defmodule ExtractInfo do
  defstruct [:title]

  @type t :: %ExtractInfo{
          title: String.t()
        }

  @callback extract(url :: String.t()) :: {:ok, ExtractInfo.t()} | {:err, nil}

  defmodule Client do
    @behaviour ExtractInfo

    require Logger

    @impl
    def extract(url) do
      with {:ok, resp} <-
             :tiddlywiki_bot
             |> Application.get_env(:extract_info_url)
             |> Mojito.post([{"content-type", "application/json"}], Jason.encode!(%{url: url})),
           %{status_code: 200, body: body} <-
             resp,
           {:ok, %{"title" => title}} <-
             Jason.decode(body) do
        {:ok, %ExtractInfo{title: title}}
      else
        err ->
          Logger.error("Cannot connect to extract info service #{inspect(err)}")
          {:error, err}
      end
    end
  end
end
