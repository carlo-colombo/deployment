defmodule Common.Reconnect do
  use GenServer

  def init(init_arg), do: {:ok, init_arg}

  def start_link(supervisor) do
    Task.start_link(__MODULE__, :run, [supervisor])
  end

  def run(supervisor) do
    receive do
    after
      5 * 60_000 ->
        Supervisor.restart_child(supervisor, :connection)
        run(supervisor)
    end
  end
end
