defmodule Dashboard.HearhbeatTest do
  use Dashboard.DataCase

  alias Dashboard.Hearhbeat

  describe "pings" do
    alias Dashboard.Hearhbeat.Ping

    import Dashboard.HearhbeatFixtures

    @invalid_attrs %{}

    test "list_pings/0 returns all pings" do
      ping = ping_fixture()
      assert Hearhbeat.list_pings() == [ping]
    end

    test "get_ping!/1 returns the ping with given id" do
      ping = ping_fixture()
      assert Hearhbeat.get_ping!(ping.id) == ping
    end

    test "create_ping/1 with valid data creates a ping" do
      valid_attrs = %{}

      assert {:ok, %Ping{} = ping} = Hearhbeat.create_ping(valid_attrs)
    end

    test "create_ping/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = Hearhbeat.create_ping(@invalid_attrs)
    end

    test "update_ping/2 with valid data updates the ping" do
      ping = ping_fixture()
      update_attrs = %{}

      assert {:ok, %Ping{} = ping} = Hearhbeat.update_ping(ping, update_attrs)
    end

    test "update_ping/2 with invalid data returns error changeset" do
      ping = ping_fixture()
      assert {:error, %Ecto.Changeset{}} = Hearhbeat.update_ping(ping, @invalid_attrs)
      assert ping == Hearhbeat.get_ping!(ping.id)
    end

    test "delete_ping/1 deletes the ping" do
      ping = ping_fixture()
      assert {:ok, %Ping{}} = Hearhbeat.delete_ping(ping)
      assert_raise Ecto.NoResultsError, fn -> Hearhbeat.get_ping!(ping.id) end
    end

    test "change_ping/1 returns a ping changeset" do
      ping = ping_fixture()
      assert %Ecto.Changeset{} = Hearhbeat.change_ping(ping)
    end
  end

  describe "pings" do
    alias Dashboard.Hearhbeat.Ping

    import Dashboard.HearhbeatFixtures

    @invalid_attrs %{name: nil}

    test "list_pings/0 returns all pings" do
      ping = ping_fixture()
      assert Hearhbeat.list_pings() == [ping]
    end

    test "get_ping!/1 returns the ping with given id" do
      ping = ping_fixture()
      assert Hearhbeat.get_ping!(ping.id) == ping
    end

    test "create_ping/1 with valid data creates a ping" do
      valid_attrs = %{name: "some name"}

      assert {:ok, %Ping{} = ping} = Hearhbeat.create_ping(valid_attrs)
      assert ping.name == "some name"
    end

    test "create_ping/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = Hearhbeat.create_ping(@invalid_attrs)
    end

    test "update_ping/2 with valid data updates the ping" do
      ping = ping_fixture()
      update_attrs = %{name: "some updated name"}

      assert {:ok, %Ping{} = ping} = Hearhbeat.update_ping(ping, update_attrs)
      assert ping.name == "some updated name"
    end

    test "update_ping/2 with invalid data returns error changeset" do
      ping = ping_fixture()
      assert {:error, %Ecto.Changeset{}} = Hearhbeat.update_ping(ping, @invalid_attrs)
      assert ping == Hearhbeat.get_ping!(ping.id)
    end

    test "delete_ping/1 deletes the ping" do
      ping = ping_fixture()
      assert {:ok, %Ping{}} = Hearhbeat.delete_ping(ping)
      assert_raise Ecto.NoResultsError, fn -> Hearhbeat.get_ping!(ping.id) end
    end

    test "change_ping/1 returns a ping changeset" do
      ping = ping_fixture()
      assert %Ecto.Changeset{} = Hearhbeat.change_ping(ping)
    end
  end

  describe "alarms" do
    alias Dashboard.Hearhbeat.Alarm

    import Dashboard.HearhbeatFixtures

    @invalid_attrs %{name: nil}

    test "list_alarms/0 returns all alarms" do
      alarm = alarm_fixture()
      assert Hearhbeat.list_alarms() == [alarm]
    end

    test "get_alarm!/1 returns the alarm with given id" do
      alarm = alarm_fixture()
      assert Hearhbeat.get_alarm!(alarm.id) == alarm
    end

    test "create_alarm/1 with valid data creates a alarm" do
      valid_attrs = %{name: "some name"}

      assert {:ok, %Alarm{} = alarm} = Hearhbeat.create_alarm(valid_attrs)
      assert alarm.name == "some name"
    end

    test "create_alarm/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = Hearhbeat.create_alarm(@invalid_attrs)
    end

    test "update_alarm/2 with valid data updates the alarm" do
      alarm = alarm_fixture()
      update_attrs = %{name: "some updated name"}

      assert {:ok, %Alarm{} = alarm} = Hearhbeat.update_alarm(alarm, update_attrs)
      assert alarm.name == "some updated name"
    end

    test "update_alarm/2 with invalid data returns error changeset" do
      alarm = alarm_fixture()
      assert {:error, %Ecto.Changeset{}} = Hearhbeat.update_alarm(alarm, @invalid_attrs)
      assert alarm == Hearhbeat.get_alarm!(alarm.id)
    end

    test "delete_alarm/1 deletes the alarm" do
      alarm = alarm_fixture()
      assert {:ok, %Alarm{}} = Hearhbeat.delete_alarm(alarm)
      assert_raise Ecto.NoResultsError, fn -> Hearhbeat.get_alarm!(alarm.id) end
    end

    test "change_alarm/1 returns a alarm changeset" do
      alarm = alarm_fixture()
      assert %Ecto.Changeset{} = Hearhbeat.change_alarm(alarm)
    end
  end
end
