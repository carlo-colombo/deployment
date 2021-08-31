defmodule Dashboard.Release do
  use GenServer
  @app :dashboard

  require Logger

  def init(init_args) do
    setup_db

    {:ok, init_args}
  end

  def start_link(_) do
    GenServer.start_link(__MODULE__, %{})
  end

  def setup_db do
    load_app()

    for repo <- repos() do
      {:ok, _, _} =
        Ecto.Migrator.with_repo(repo, fn repo ->
          case repo.__adapter__.storage_up(repo.config()) do
            :ok ->
              :ok

            {:error, :already_up} ->
              Logger.info("Storage already up")
              :ok

            e ->
              e
          end
        end)

      {:ok, _, _} = Ecto.Migrator.with_repo(repo, &Ecto.Migrator.run(&1, :up, all: true))
    end
  end

  def rollback(repo, version) do
    load_app()
    {:ok, _, _} = Ecto.Migrator.with_repo(repo, &Ecto.Migrator.run(&1, :down, to: version))
  end

  defp repos do
    Application.fetch_env!(@app, :ecto_repos)
  end

  defp load_app do
    Application.load(@app)
  end
end
