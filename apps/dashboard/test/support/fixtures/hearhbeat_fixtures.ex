defmodule Dashboard.HearhbeatFixtures do
  @moduledoc """
  This module defines test helpers for creating
  entities via the `Dashboard.Hearhbeat` context.
  """

  @doc """
  Generate a ping.
  """
  def ping_fixture(attrs \\ %{}) do
    {:ok, ping} =
      attrs
      |> Enum.into(%{

      })
      |> Dashboard.Hearhbeat.create_ping()

    ping
  end

  @doc """
  Generate a ping.
  """
  def ping_fixture(attrs \\ %{}) do
    {:ok, ping} =
      attrs
      |> Enum.into(%{
        name: "some name"
      })
      |> Dashboard.Hearhbeat.create_ping()

    ping
  end
end
