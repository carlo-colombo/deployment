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

config :nadia,
  token: System.get_env("TELEGRAM_BOT_TOKEN")

config :tiddlywiki_bot,
  extract_info_url: System.get_env("EXTRACT_INFO_URL"),
  tasks_filter: "[tag[chandler]!tag[done]]",
  register: System.get_env("REGISTER_ADDRESS", "https://api.telegram.org"),
  own: System.get_env("OWN_ADDRESS"),
  wiki: [
    url: System.get_env("WIKI_URL"),
    username: System.get_env("WIKI_USERNAME"),
    password: System.get_env("WIKI_PASSWORD"),
    external_url: System.get_env("WIKI_EXTERNAL_URL")
  ]
