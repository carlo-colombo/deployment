defmodule Feed2wiki do
  @moduledoc """
  Documentation for `Feed2wiki`.
  """

  require Logger

  use GenServer

  def init(init_arg), do: {:ok, init_arg}

  def start_link(_arg) do
    Task.start_link(__MODULE__, :run, [])
  end

  def run do
    collect_feed()
    receive do
    after
      5 * 60_000 -> :ok
    end
  end

  defp bot() do
    {:ok, hostname} = :inet.gethostname()

    [
      Application.get_env(:feed2wiki, :bot_name),
      hostname
    ]
    |> Enum.join("@")
    |> String.to_atom()
  end

  def collect_feed() do
    Logger.info("Collecting...")

    with {:ok, tiddler} <- :rpc.call(bot(), TiddlyWiki, :get, ["rssconfig"]),
         %{text: text, type: "application/json"} <- tiddler,
         {:ok, urls} <- Jason.decode(text) do
      urls
      |> Enum.map(&get_feed/1)
      |> Enum.each(fn t ->
        Logger.info("adding item: #{inspect(t)}")
      end)
    else
      err ->
        Logger.error("cannot read rssconfig: #{inspect(err)}")
    end
  end

  defp get_feed(url) do
    with {:ok, resp} <- Mojito.get(url),
         %{status_code: 200, body: body} <- resp,
         {:ok, feed, _} <- FeederEx.parse(body) do
      {:ok, _} =
        feed
        |> Map.get(:entries)
        |> List.first()
        |> entry_to_tiddler(feed)
        |> put
    else
      err ->
        Logger.error("cannot post entry from feed #{url}: #{inspect(err)}")
        err
    end
  end

  defp put(%{title: title}=tiddler) do
    case :rpc.call(bot(), TiddlyWiki, :get, [title]) do
      {:ok, _} ->
        Logger.info("Entry already exists, skipping (#{title})")
        {:ok, :skipped}
      {:error, %{status_code: 404}}->
        :rpc.call(bot(), TiddlyWiki, :put, [tiddler])
      e ->
        Loggger.info("unexpected")
        {:ok, e}
    end
  end

  defp entry_to_tiddler(entry, feed) do
    :rpc.call(bot(), TiddlyWiki.Tiddler, :__struct__, [
      %{
        title: "[feed] " <> entry.title,
        text: "" <> entry.summary <> "<br>" <> entry.link,
        tags: "feed " <> feed.link
      }
    ])
  end
end
