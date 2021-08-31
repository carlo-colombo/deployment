defmodule DashboardWeb.AlarmControllerTest do
  use DashboardWeb.ConnCase

  import Dashboard.HearhbeatFixtures

  alias Dashboard.Hearhbeat.Alarm

  @create_attrs %{
    name: "some name"
  }
  @update_attrs %{
    name: "some updated name"
  }
  @invalid_attrs %{name: nil}

  setup %{conn: conn} do
    {:ok, conn: put_req_header(conn, "accept", "application/json")}
  end

  describe "index" do
    test "lists all alarms", %{conn: conn} do
      conn = get(conn, Routes.alarm_path(conn, :index))
      assert json_response(conn, 200)["data"] == []
    end
  end

  describe "create alarm" do
    test "renders alarm when data is valid", %{conn: conn} do
      conn = post(conn, Routes.alarm_path(conn, :create), alarm: @create_attrs)
      assert %{"id" => id} = json_response(conn, 201)["data"]

      conn = get(conn, Routes.alarm_path(conn, :show, id))

      assert %{
               "id" => ^id,
               "name" => "some name"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn} do
      conn = post(conn, Routes.alarm_path(conn, :create), alarm: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "update alarm" do
    setup [:create_alarm]

    test "renders alarm when data is valid", %{conn: conn, alarm: %Alarm{id: id} = alarm} do
      conn = put(conn, Routes.alarm_path(conn, :update, alarm), alarm: @update_attrs)
      assert %{"id" => ^id} = json_response(conn, 200)["data"]

      conn = get(conn, Routes.alarm_path(conn, :show, id))

      assert %{
               "id" => ^id,
               "name" => "some updated name"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn, alarm: alarm} do
      conn = put(conn, Routes.alarm_path(conn, :update, alarm), alarm: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "delete alarm" do
    setup [:create_alarm]

    test "deletes chosen alarm", %{conn: conn, alarm: alarm} do
      conn = delete(conn, Routes.alarm_path(conn, :delete, alarm))
      assert response(conn, 204)

      assert_error_sent 404, fn ->
        get(conn, Routes.alarm_path(conn, :show, alarm))
      end
    end
  end

  defp create_alarm(_) do
    alarm = alarm_fixture()
    %{alarm: alarm}
  end
end
