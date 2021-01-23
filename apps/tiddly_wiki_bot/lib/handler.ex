defmodule TiddlyWikiBot.Handler do
  require Logger

  import Plug.Conn



  @doc ~S"""
  iex> "_*[]()~`>#+-=|{}.!"
  ...> |> TiddlyWikiBot.Handler.escape_markdown_v2
  "\\_\\*\\[\\]\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!"
  """
  def escape_markdown_v2(s) do
    ~r/([\_\*\[\]\(\)\~\`\>\#\+\-\=\|\{\}\.\!])/
    |> Regex.replace(s, "\\\\\\1")
  end

  use Agent

  def start_link(_) do
    Logger.info("starting #{__MODULE__}")
    Agent.start_link(fn -> %{} end, name: __MODULE__)
  end

  @doc ~S"""
      iex> TiddlyWikiBot.Handler.insert(12, "12")
      :ok
      iex> Agent.update(TiddlyWikiBot.Handler, &Map.delete(&1, 12))
  """
  def insert(chat_id, tiddler) do
    Agent.update(__MODULE__, fn map -> Map.update(map, chat_id, [tiddler], &[tiddler | &1]) end)
  end

  @doc ~S"""
      iex> TiddlyWikiBot.Handler.insert(12, "14")
      :ok
      iex> TiddlyWikiBot.Handler.insert(12, "18")
      :ok
      iex> TiddlyWikiBot.Handler.get(12)
      ["18", "14"]
      iex> Agent.update(TiddlyWikiBot.Handler, &Map.delete(&1, 12))
  """
  def get(chat_id) do
    Agent.get(__MODULE__, &Map.get(&1, chat_id))
  end

  def new_tiddler(title, text, chat_id, n \\ 0)

  def new_tiddler(title, text, chat_id, n) when n < 5 do
    case TiddlyWiki.get(title) do
      {:ok, _} ->
        new_tiddler(title <> " - dup", text, chat_id, n + 1)

      {:error, %{status_code: 404}} ->
        %TiddlyWiki.Tiddler{
          title: title,
          tags: "external",
          text: text,
          fields: %{
            chat_id: chat_id
          }
        }
    end
  end

  def new_tiddler(_, _, _, _), do: nil

  def get_tg_client, do: Application.get_env(:tiddlywiki_bot, :telegram_client)
  def get_tw_client, do: Application.get_env(:tiddlywiki_bot, :tiddlywiki_client)
  def get_ei_client, do: Application.get_env(:tiddlywiki_bot, :extract_info_client)

  def init(options), do: options

  def call(%{body_params: update} = conn, opts) do
    %{
      "message" =>
        %{
          "chat" => %{
            "id" => chat_id
          },
          "text" => text
        } = message
    } = update

    Logger.info("dispatching #{text}.")

    module =
      cond do
        match?(%{"text" => "/tasks"}, message) ->
          TiddlyWikiBot.Handler.Tasks

        match?(%{"text" => "/create" <> _}, message) ->
          TiddlyWikiBot.Handler.Create

        match?(%{"document" => _}, message) || match?(%{"photo" => _}, message) ->
          TiddlyWikiBot.Handler.Upload

        true ->
          TiddlyWikiBot.Handler.Link
      end

    conn
    |> assign(:chat_id, chat_id)
    |> assign_text(message)
    |> module.call(module.init(opts))
    |> put_resp_header("content-type", "application/json")
    |> send_resp(201, "{}")
  rescue
    e ->
      error_message = "Something failed dispatching #{inspect(conn)}. Error was #{inspect(e)}"
      Logger.error(error_message)
      send_resp(conn, 200, error_message)
  end

  def list_commands do
    []
  end

  defp assign_text(conn, %{"text" => text}), do: assign(conn, :text, text)
  defp assign_text(conn, _), do: conn
end
