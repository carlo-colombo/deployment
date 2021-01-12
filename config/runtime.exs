import Config

{:ok, hostname} = :inet.gethostname()

config :libcluster, topologies: [
  connection: [
    strategy: Elixir.Cluster.Strategy.Epmd,
    config: [
      hosts:
      ["dashboard", "tiddlywiki_bot", "feed2wiki"]
      |> Enum.map(&Enum.join([&1, hostname], "@"))
      |> Enum.map(&String.to_atom/1)
    ]
  ]
]

config :feed2wiki, bot_name: "tiddlywiki_bot"

