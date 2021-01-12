defmodule TiddlyWikiBot.Router do
  use Plug.Router

  plug(Plug.Logger)

  plug(Plug.Parsers,
    parsers: [:urlencoded, :json],
    pass: ["application/json"],
    json_decoder: Jason
  )

  plug(:match)
  plug(TokenValidation)
  plug(:as_session_id)
  plug(Plug.Session, store: TiddlyWikiBot.ETS, key: "chat_id", table: :session)
  plug(:fetch)
  plug(:dispatch)

  post("/api/:token/hook", to: TiddlyWikiBot.Handler)

  match _ do
    send_resp(conn, 404, "oops")
  end

  defp fetch(conn, _opts), do: fetch_session(conn)

  defp as_session_id(%{body_params: %{"message" => %{"chat" => %{"id" => id}}}} = conn, _opts) do
    conn
    |> assign(:chat_id, id)
    |> put_req_header("cookie", "chat_id=#{id}")
  end

  defp as_session_id(conn, _opts), do: conn
end
