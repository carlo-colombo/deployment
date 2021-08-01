defmodule DashboardWeb.PingView do
  use DashboardWeb, :view
  alias DashboardWeb.PingView

  def render("index.json", %{pings: pings}) do
    %{data: render_many(pings, PingView, "ping.json")}
  end

  def render("show.json", %{ping: ping}) do
    %{data: render_one(ping, PingView, "ping.json")}
  end

  def render("ping.json", %{ping: ping}) do
    %{
      id: ping.id,
      name: ping.name,
      inserted_at: ping.inserted_at
    }
  end
end
