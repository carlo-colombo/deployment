defmodule DashboardWeb.PingController do
  use DashboardWeb, :controller

  alias Dashboard.Hearhbeat
  alias Dashboard.Hearhbeat.Ping

  action_fallback DashboardWeb.FallbackController

  def index(conn, _params) do
    pings = Hearhbeat.list_pings()
    render(conn, "index.json", pings: pings)
  end

  def create(conn, params) do
    with {:ok, %Ping{} = ping} <- Hearhbeat.create_ping(params) do
      conn
      |> put_status(:created)
      |> put_resp_header("location", Routes.ping_path(conn, :show, ping))
      |> render("show.json", ping: ping)
    end
  end

  def show(conn, %{"id" => id}) do
    ping = Hearhbeat.get_ping!(id)
    render(conn, "show.json", ping: ping)
  end

  def update(conn, %{"id" => id, "ping" => ping_params}) do
    ping = Hearhbeat.get_ping!(id)

    with {:ok, %Ping{} = ping} <- Hearhbeat.update_ping(ping, ping_params) do
      render(conn, "show.json", ping: ping)
    end
  end

  def delete(conn, %{"id" => id}) do
    ping = Hearhbeat.get_ping!(id)

    with {:ok, %Ping{}} <- Hearhbeat.delete_ping(ping) do
      send_resp(conn, :no_content, "")
    end
  end
end
