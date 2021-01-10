require Protocol

Protocol.derive(Jason.Encoder, Commander.Command, only: [:command, :description])

defmodule TiddlyWikiBot.Register do
  use GenServer
  require Logger

  def init(init_arg), do: {:ok, init_arg}

  def start_link(:test), do: GenServer.start_link(__MODULE__, [])

  def start_link(_arg) do
    register()
    set_commands()
    Task.start_link(__MODULE__, :run, [])
  end

  defp set_commands do
    cmds = TiddlyWikiBot.Handler.list_commands()

    Logger.info("Setting up commands: #{inspect(cmds)}")

    cmds
    |> Jason.encode!()
    |> (&Nadia.API.request("setMyCommands", commands: &1)).()
  end

  defp register do
    own = Application.get_env(:tiddlywiki_bot, :own)
    register = Application.get_env(:tiddlywiki_bot, :register)
    token = Application.get_env(:nadia, :token)

    Logger.info("Registering '#{own}' on register '#{register}'")

    with {:ok, resp} <-
           register
           |> URI.merge("/bot#{token}/setWebhook")
           |> Map.put(:query, URI.encode_query(%{"url" => own}))
           |> Mojito.get(),
         %{status_code: 200, body: body} <- resp,
         {:ok, decoded} <- Jason.decode(body) do
      Logger.info("#{own} registered: #{inspect(decoded)}")
    else
      error -> Logger.error("Registration failed: #{inspect(error)}")
    end
  end

  def dontrun, do: nil

  def run do
    receive do
    after
      5 * 60_000 ->
        register()
        run()
    end
  end
end
