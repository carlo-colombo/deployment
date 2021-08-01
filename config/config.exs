# This file is responsible for configuring your application
# and its dependencies with the aid of the Mix.Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.

# General application configuration
use Mix.Config

# Configures the endpoint

config :dashboard,
  ecto_repos: [Dashboard.Repo]

# Configures the endpoint
config :dashboard, DashboardWeb.Endpoint,
  url: [host: "localhost"],
  secret_key_base: "3sn9TM0ubIcOSSdWJ6n+fxIST9yjamp1cdDxLJTbkSQKAQKZOJOZX4Rm9Hh7HJ7l",
  render_errors: [view: DashboardWeb.ErrorView, accepts: ~w(html json), layout: false],
  pubsub_server: Dashboard.PubSub,
  live_view: [signing_salt: "hPmnWLbX"]

# Configure esbuild (the version is required)
config :esbuild,
  version: "0.12.17",
  default: [
    args: ~w(js/app.js --bundle --target=es2016 --outdir=../priv/static/assets),
    cd: Path.expand("../apps/dashboard/assets", __DIR__),
    env: %{"NODE_PATH" => Path.expand("../deps", __DIR__)}
  ]

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Use Jason for JSON parsing in Phoenix
config :phoenix, :json_library, Jason

config :tiddlywiki_bot,
  telegram_client: Nadia,
  tiddlywiki_client: TiddlyWiki.Client,
  extract_info_client: ExtractInfo.Client,
  tasks_filter: "[tag[chandler]!tag[done]]"

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Use Jason for JSON parsing in Phoenix
config :phoenix, :json_library, Jason

config :logger, :console,
  format: "[$level] $message $metadata\n",
  metadata: [:mfa]

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.

# Configure esbuild (the version is required)
# Configures the mailer.
#
# By default it uses the "Local" adapter which stores the emails
# locally. You can see the emails in your browser, at "/dev/mailbox".
#
# For production it's recommended to configure a different adapter
# at the `config/runtime.exs`.

import_config "#{Mix.env()}.exs"
