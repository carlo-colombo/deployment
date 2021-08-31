defmodule Dashboard.Repo do
  @app :dashboard

  use Ecto.Repo,
    otp_app: @app,
    adapter: Ecto.Adapters.SQLite3

  # @on_load :migrate
end
