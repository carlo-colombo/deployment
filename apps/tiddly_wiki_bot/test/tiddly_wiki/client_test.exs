defmodule TiddlyWiki.ClientTest do
  use ExUnit.Case, async: true

  alias TiddlyWiki.Client
  alias TiddlyWiki.Tiddler

  setup do
    bypass = Bypass.open()
    {:ok, bypass: bypass}
  end

  describe "put" do
    test "create tiddler", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        assert "/recipes/default/tiddlers/test%3A%20tiddler%2F%2Ffoo" == conn.request_path
        assert "PUT" == conn.method

        conn
        |> check_headers

        {:ok, body, conn} = Plug.Conn.read_body(conn)

        assert %Tiddler{
                 title: "test: tiddler//foo",
                 tags: "external",
                 creator: "foo",
                 modifier: "foo"
               } == struct(Tiddler, Jason.decode!(body, keys: :atoms))

        Plug.Conn.resp(conn, 204, "")
      end)

      assert {:ok, "http://ext.example.com#test%3A%20tiddler%2F%2Ffoo"} ==
               Client.put(
                 %TiddlyWiki{
                   url: endpoint_url(bypass.port),
                   external_url: "http://ext.example.com",
                   username: "foo",
                   password: "bar"
                 },
                 %Tiddler{title: "test: tiddler//foo", tags: "external"}
               )
    end

    test "handles invalid status codes (not 204)", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        Plug.Conn.resp(conn, 401, "")
      end)

      assert {:error, _} =
               Client.put(
                 %TiddlyWiki{
                   url: endpoint_url(bypass.port),
                   external_url: "http://ext.example.com",
                   username: "foo",
                   password: "bar"
                 },
                 %Tiddler{title: "test tiddler", tags: "external"}
               )
    end
  end

  describe "get_all" do
    test "client handles requests", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        assert "/recipes/default/tiddlers.json" == conn.request_path
        assert "GET" == conn.method

        Plug.Conn.resp(conn, 200, ~s<[{"title":"foobar"}]>)
      end)

      assert {
               :ok,
               [struct(Tiddler, title: "foobar")]
             } == Client.get_all(%TiddlyWiki{url: endpoint_url(bypass.port)})
    end

    test "client send authentication headers", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        assert "/recipes/default/tiddlers.json" == conn.request_path
        assert "GET" == conn.method

        conn
        |> check_headers

        Plug.Conn.resp(conn, 200, ~s<[{"title":"foobar"}]>)
      end)

      assert {:ok, _} =
               Client.get_all(%TiddlyWiki{
                 url: endpoint_url(bypass.port),
                 username: "foo",
                 password: "bar"
               })
    end

    test "filter is encoded", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        assert "filter=%5B%21tag%5Bfoo%5D%5D" == conn.query_string

        Plug.Conn.resp(conn, 200, ~s<[{"title":"foobar"}]>)
      end)

      {:ok, _} =
        Client.get_all(
          %TiddlyWiki{
            url: endpoint_url(bypass.port)
          },
          "[!tag[foo]]"
        )
    end

    test "handles invalid status codes (not 200)", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        Plug.Conn.resp(conn, 401, "")
      end)

      assert {:error, _} =
               Client.get_all(%TiddlyWiki{
                 url: endpoint_url(bypass.port)
               })
    end
  end

  describe "get" do
    test "retrieves a tiddler", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        assert "/recipes/default/tiddlers/a%2F%2F%3Ab" == conn.request_path
        assert "GET" == conn.method

        conn
        |> check_headers

        Plug.Conn.resp(conn, 200, ~s<{"title":"a//:b"}>)
      end)

      assert {:ok, %Tiddler{title: "a//:b"}} ==
               Client.get(
                 %TiddlyWiki{
                   url: endpoint_url(bypass.port),
                   username: "foo",
                   password: "bar"
                 },
                 "a//:b"
               )
    end

    test "handles invalid status codes (not 200)", %{bypass: bypass} do
      Bypass.expect(bypass, fn conn ->
        Plug.Conn.resp(conn, 404, "")
      end)

      assert {:error, %{status_code: 404}} ==
               Client.get(
                 %TiddlyWiki{
                   url: endpoint_url(bypass.port)
                 },
                 "not a title"
               )
    end
  end

  defp endpoint_url(port), do: "http://localhost:#{port}"

  def check_headers(conn) do
    conn.req_headers
    |> Enum.each(fn
      {"authorization", v} -> assert "Basic #{"foo:bar" |> Base.encode64()}" == v
      {"X-Requested-With", v} -> assert "TiddlyWiki" == v
      _ -> :ok
    end)
  end
end
