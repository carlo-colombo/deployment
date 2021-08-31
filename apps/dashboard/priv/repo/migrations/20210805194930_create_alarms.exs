defmodule Dashboard.Repo.Migrations.CreateAlarms do
  use Ecto.Migration

  def change do
    create table(:alarms) do
      add :name, :string

      timestamps()
    end
  end
end
