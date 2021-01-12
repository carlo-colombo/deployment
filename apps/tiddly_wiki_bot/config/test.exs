use Mix.Config

config :nadia,
  token: "a_valid_token"

config :logger,
  level: :error

config :tiddlywiki_bot,
  tasks_filter: "[tag[foo]]"
