defmodule TiddlyWikiBot.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  def start(_type, [mix_env]) do
    :ets.new(:session, [:named_table, :public, read_concurrency: true])
    # List all child processes to be supervised
    children = [
      # Starts a worker by calling: GatherInfoTelegramBot.Worker.start_link(arg)
      # {GatherInfoTelegramBot.Worker, arg},
      Plug.Cowboy.child_spec(
        scheme: :http,
        plug: TiddlyWikiBot.Router,
        options: [port: 9021]
      ),
      {TiddlyWikiBot.Register, mix_env},
      TiddlyWikiBot.Handler,
      {Cluster.Supervisor,
       [Application.get_env(:libcluster, :topologies), [name: TiddlyWikiBot.ClusterSupervisor]]},
      {Common.Reconnect, TiddlyWikiBot.ClusterSupervisor}
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: TiddlyWikiBot.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
