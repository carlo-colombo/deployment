defmodule TokenValidation do
  use Plug.Builder

  plug(:validate)

  def validate(%{params: params} = conn, _opts) do
    token = Map.get(params, "token")

    if token != Nadia.Config.token() do
      conn
      |> send_resp(501, "Invalid token '#{token}'")
      |> halt
    else
      conn
    end
  end
end
