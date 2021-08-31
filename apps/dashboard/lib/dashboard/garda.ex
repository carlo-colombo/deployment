require Protocol

defmodule Dashboard.Garda do
  use GenServer
  require Logger

  alias Dashboard.Hearhbeat

  def init(init_arg), do: {:ok, init_arg}

  def start_link(:test), do: GenServer.start_link(__MODULE__, [])

  def start_link(_arg) do
    Task.start_link(__MODULE__, :run, [])
  end

  def check() do
    for name <- Hearhbeat.names() do
      Logger.info("Checking: #{name}")

      if !Hearhbeat.recent_pings_without_alarm("-1 hours", "-5 minutes", name) do
        {:ok, _} =
          Nadia.send_message(System.get_env("DESTINATION_CHAT_ID"), "Warn! Warn! #{name}")

        Hearhbeat.create_alarm(%{name: name})
      else
        Logger.info("Nothing to see here")
      end
    end
  end

  def dontrun, do: nil

  def run do
    receive do
    after
      60_000 ->
        check()
        run()
    end
  end
end
