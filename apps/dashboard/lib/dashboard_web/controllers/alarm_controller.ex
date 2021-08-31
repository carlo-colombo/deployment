defmodule DashboardWeb.AlarmController do
  use DashboardWeb, :controller

  alias Dashboard.Hearhbeat
  alias Dashboard.Hearhbeat.Alarm

  action_fallback DashboardWeb.FallbackController

  def index(conn, _params) do
    alarms = Hearhbeat.list_alarms()
    render(conn, "index.json", alarms: alarms)
  end

  def create(conn, %{"alarm" => alarm_params}) do
    with {:ok, %Alarm{} = alarm} <- Hearhbeat.create_alarm(alarm_params) do
      conn
      |> put_status(:created)
      |> put_resp_header("location", Routes.alarm_path(conn, :show, alarm))
      |> render("show.json", alarm: alarm)
    end
  end

  def show(conn, %{"id" => id}) do
    alarm = Hearhbeat.get_alarm!(id)
    render(conn, "show.json", alarm: alarm)
  end

  def update(conn, %{"id" => id, "alarm" => alarm_params}) do
    alarm = Hearhbeat.get_alarm!(id)

    with {:ok, %Alarm{} = alarm} <- Hearhbeat.update_alarm(alarm, alarm_params) do
      render(conn, "show.json", alarm: alarm)
    end
  end

  def delete(conn, %{"id" => id}) do
    alarm = Hearhbeat.get_alarm!(id)

    with {:ok, %Alarm{}} <- Hearhbeat.delete_alarm(alarm) do
      send_resp(conn, :no_content, "")
    end
  end
end
