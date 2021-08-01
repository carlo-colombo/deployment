defmodule Dashboard.Hearhbeat do
  @moduledoc """
  The Hearhbeat context.
  """

  import Ecto.Query, warn: false
  alias Dashboard.Repo

  alias Dashboard.Hearhbeat.Ping

  @doc """
  Returns the list of pings.

  ## Examples

      iex> list_pings()
      [%Ping{}, ...]

  """
  def list_pings do
    Repo.all(Ping)
  end

  @doc """
  Gets a single ping.

  Raises `Ecto.NoResultsError` if the Ping does not exist.

  ## Examples

      iex> get_ping!(123)
      %Ping{}

      iex> get_ping!(456)
      ** (Ecto.NoResultsError)

  """
  def get_ping!(id), do: Repo.get!(Ping, id)

  @doc """
  Creates a ping.

  ## Examples

      iex> create_ping(%{field: value})
      {:ok, %Ping{}}

      iex> create_ping(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_ping(attrs \\ %{}) do
    %Ping{}
    |> Ping.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a ping.

  ## Examples

      iex> update_ping(ping, %{field: new_value})
      {:ok, %Ping{}}

      iex> update_ping(ping, %{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def update_ping(%Ping{} = ping, attrs) do
    ping
    |> Ping.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a ping.

  ## Examples

      iex> delete_ping(ping)
      {:ok, %Ping{}}

      iex> delete_ping(ping)
      {:error, %Ecto.Changeset{}}

  """
  def delete_ping(%Ping{} = ping) do
    Repo.delete(ping)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking ping changes.

  ## Examples

      iex> change_ping(ping)
      %Ecto.Changeset{data: %Ping{}}

  """
  def change_ping(%Ping{} = ping, attrs \\ %{}) do
    Ping.changeset(ping, attrs)
  end
end
