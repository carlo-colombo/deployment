defmodule TiddlyWikiBot.ETS do
  @behaviour Plug.Session.Store
  alias Plug.Session.ETS

  defdelegate init(opts), to: ETS

  def get(conn, sid, table) do
    {_, data} = ETS.get(conn, sid, table)
    {sid, data}
  end

  defdelegate put(conn, sid, data, table), to: ETS
  defdelegate delete(conn, sid, table), to: ETS
end
