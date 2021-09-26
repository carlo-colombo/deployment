defmodule DashboardWeb.PageController do
  use DashboardWeb, :live_view

  alias Dashboard.Hearhbeat

  # def index(conn, _params) do
  #   render(conn, "index.html")
  # end
  #
  def render(assigns) do
    ~H"""
    <ul>
    <%= for {name, last_ping, last_alarm} <- @last_pings do %>
      <li><%= name %>: <%= last_ping %> / <%= last_alarm %></li>
    <% end %>
    </ul>
    """
  end

  def mount(params, _, socket) do
    if connected?(socket), do: Process.send_after(self(), :update, 30000)

    IO.inspect(params)
    last_pings = Hearhbeat.last_pings()
    {:ok, assign(socket, :last_pings, last_pings)}
  end

  def handle_info(:update, socket) do
    Process.send_after(self(), :update, 30000)
    last_pings = Hearhbeat.last_pings()
    {:noreply, assign(socket, :last_pings, last_pings)}
  end

  def live(), do: nil
end
