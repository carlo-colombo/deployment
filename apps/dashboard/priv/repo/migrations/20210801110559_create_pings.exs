defmodule Dashboard.Repo.Migrations.CreatePings do
  use Ecto.Migration

  def change do
    create table(:pings) do
      add :name, :string

      timestamps()
    end
  end
end
