defmodule Dashboard.Hearhbeat do
  @moduledoc """
  The Hearhbeat context.
  """

  import Ecto.Query, warn: false
  alias Dashboard.Repo

  alias Dashboard.Hearhbeat.Ping
  alias Dashboard.Hearhbeat.Alarm

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

  def alarms_recent_than(time_ago) do
    from(a in Alarm,
      where:
        fragment(
          "strftime('%s', inserted_at) > strftime('%s','now',?)",
          ^time_ago
        )
    )
    |> Repo.all()
  end

  def names() do
    from(p in Ping, select: p.name)
    |> Repo.all()
    |> Enum.uniq()
  end

  def recent_pings_without_alarm(a, b, name) do
    recent_ping =
      from(
        p in Ping,
        where:
          p.name == ^name and
            fragment("Strftime('%s', ?) > Strftime('%s', 'now', ?)", p.inserted_at, ^b)
      )

    recent_alarm =
      from(
        a in Alarm,
        where:
          a.name == ^name and
            fragment("Strftime('%s', ?) > Strftime('%s', 'now',   ?)", a.inserted_at, ^a)
      )

    recent_ping
    |> union(^recent_alarm)
    |> Repo.exists?()
  end

  def last_pings do
    from(p in Ping,
      left_join: a in Alarm,
      on: a.name == p.name,
      select: {p.name, max(p.inserted_at), max(a.inserted_at)},
      group_by: p.name,
      order_by: p.name
    )
    |> Repo.all()
  end

  alias Dashboard.Hearhbeat.Alarm

  @doc """
  Returns the list of alarms.

  ## Examples

      iex> list_alarms()
      [%Alarm{}, ...]

  """
  def list_alarms do
    Repo.all(Alarm)
  end

  @doc """
  Gets a single alarm.

  Raises `Ecto.NoResultsError` if the Alarm does not exist.

  ## Examples

      iex> get_alarm!(123)
      %Alarm{}

      iex> get_alarm!(456)
      ** (Ecto.NoResultsError)

  """
  def get_alarm!(id), do: Repo.get!(Alarm, id)

  @doc """
  Creates a alarm.

  ## Examples

      iex> create_alarm(%{field: value})
      {:ok, %Alarm{}}

      iex> create_alarm(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_alarm(attrs \\ %{}) do
    %Alarm{}
    |> Alarm.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a alarm.

  ## Examples

      iex> update_alarm(alarm, %{field: new_value})
      {:ok, %Alarm{}}

      iex> update_alarm(alarm, %{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def update_alarm(%Alarm{} = alarm, attrs) do
    alarm
    |> Alarm.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a alarm.

  ## Examples

      iex> delete_alarm(alarm)
      {:ok, %Alarm{}}

      iex> delete_alarm(alarm)
      {:error, %Ecto.Changeset{}}

  """
  def delete_alarm(%Alarm{} = alarm) do
    Repo.delete(alarm)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking alarm changes.

  ## Examples

      iex> change_alarm(alarm)
      %Ecto.Changeset{data: %Alarm{}}

  """
  def change_alarm(%Alarm{} = alarm, attrs \\ %{}) do
    Alarm.changeset(alarm, attrs)
  end
end
