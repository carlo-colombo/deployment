defmodule Dashboard.Hearhbeat.Ping do
  use Ecto.Schema
  import Ecto.Changeset

  schema "pings" do
    field :name, :string

    timestamps()
  end

  @doc false
  def changeset(ping, attrs) do
    ping
    |> cast(attrs, [:name])
    |> validate_required([:name])
  end
end
