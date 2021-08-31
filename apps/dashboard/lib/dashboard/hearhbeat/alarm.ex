defmodule Dashboard.Hearhbeat.Alarm do
  use Ecto.Schema
  import Ecto.Changeset

  schema "alarms" do
    field :name, :string

    timestamps()
  end

  @doc false
  def changeset(alarm, attrs) do
    alarm
    |> cast(attrs, [:name])
    |> validate_required([:name])
  end
end
