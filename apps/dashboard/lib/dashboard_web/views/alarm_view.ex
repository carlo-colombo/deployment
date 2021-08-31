defmodule DashboardWeb.AlarmView do
  use DashboardWeb, :view
  alias DashboardWeb.AlarmView

  def render("index.json", %{alarms: alarms}) do
    %{data: render_many(alarms, AlarmView, "alarm.json")}
  end

  def render("show.json", %{alarm: alarm}) do
    %{data: render_one(alarm, AlarmView, "alarm.json")}
  end

  def render("alarm.json", %{alarm: alarm}) do
    %{
      id: alarm.id,
      name: alarm.name
    }
  end
end
