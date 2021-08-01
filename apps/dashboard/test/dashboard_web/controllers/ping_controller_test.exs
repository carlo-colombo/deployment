defmodule DashboardWeb.PingControllerTest do
  use DashboardWeb.ConnCase

  import Dashboard.HearhbeatFixtures

  alias Dashboard.Hearhbeat.Ping

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
    test "lists all pings", %{conn: conn} do
      conn = get(conn, Routes.ping_path(conn, :index))
      assert json_response(conn, 200)["data"] == []
    end
  end

  describe "create ping" do
    test "renders ping when data is valid", %{conn: conn} do
      conn = post(conn, Routes.ping_path(conn, :create), ping: @create_attrs)
      assert %{"id" => id} = json_response(conn, 201)["data"]

      conn = get(conn, Routes.ping_path(conn, :show, id))

      assert %{
               "id" => ^id,
               "name" => "some name"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn} do
      conn = post(conn, Routes.ping_path(conn, :create), ping: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "update ping" do
    setup [:create_ping]

    test "renders ping when data is valid", %{conn: conn, ping: %Ping{id: id} = ping} do
      conn = put(conn, Routes.ping_path(conn, :update, ping), ping: @update_attrs)
      assert %{"id" => ^id} = json_response(conn, 200)["data"]

      conn = get(conn, Routes.ping_path(conn, :show, id))

      assert %{
               "id" => ^id,
               "name" => "some updated name"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn, ping: ping} do
      conn = put(conn, Routes.ping_path(conn, :update, ping), ping: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "delete ping" do
    setup [:create_ping]

    test "deletes chosen ping", %{conn: conn, ping: ping} do
      conn = delete(conn, Routes.ping_path(conn, :delete, ping))
      assert response(conn, 204)

      assert_error_sent 404, fn ->
        get(conn, Routes.ping_path(conn, :show, ping))
      end
    end
  end

  defp create_ping(_) do
    ping = ping_fixture()
    %{ping: ping}
  end
end
