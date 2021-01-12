defmodule Feed2wiki.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  def start(_type, [mix_env]) do
    children = [
      {Cluster.Supervisor,
       [
         Application.get_env(:libcluster, :topologies),
         [name: Feed2wiki.ClusterSupervisor]
       ]},
      Feed2wiki,
      {Common.Reconnect, Feed2wiki.ClusterSupervisor}
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: Feed2wiki.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
