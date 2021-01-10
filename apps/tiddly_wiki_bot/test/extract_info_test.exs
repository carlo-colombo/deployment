defmodule ExtractInfo.ClientTest do
  use ExUnit.Case, async: true

  alias ExtractInfo.Client

  setup do
    bypass = Bypass.open()
    Application.put_env(:tiddlywiki_bot, :extract_info_url, "http://localhost:#{bypass.port}")
    {:ok, bypass: bypass}
  end

  test "extract the title from a html page", %{bypass: bypass} do
    Bypass.expect(bypass, fn conn ->
      assert "application/json" ==
               conn.req_headers
               |> Enum.find_value(fn
                 {"content-type", v} -> v
                 _ -> false
               end)

      Plug.Conn.resp(conn, 200, ~s({"title": "the foo title"}))
    end)

    assert {
             :ok,
             %ExtractInfo{title: "the foo title"}
           } == Client.extract("http://does-not-matter.example.com")
  end

  test "returns error if service does not answer succesfully", %{bypass: bypass} do
    Bypass.expect(bypass, fn conn ->
      Plug.Conn.resp(conn, 500, "")
    end)

    assert {:error, _} = Client.extract("http://does-not-matter.example.com")
  end

  test "returns error if service does not answer" do
    Application.put_env(:tiddlywiki_bot, :extract_info_url, "http://invalid")
    assert {:error, _} = Client.extract("http://does-not-matter.example.com")
  end
end
