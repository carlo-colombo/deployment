defmodule TiddlyWikiBot.HandlerTest do
  use ExUnit.Case
  doctest TiddlyWikiBot.Handler

  import Mox
  import TestHelper

  setup :verify_on_exit!

  setup do
    bypass = Bypass.open()
    {:ok, bypass: bypass}
  end

  alias TiddlyWiki.Tiddler

  setup_all do
    Application.put_env(:tiddlywiki_bot, :tiddlywiki_client, TiddlyWikiMock)
    Application.put_env(:tiddlywiki_bot, :telegram_client, NadiaMock)
    Application.put_env(:tiddlywiki_bot, :extract_info_client, ExtractInfoMock)

    Application.put_env(:tiddlywiki_bot, :wiki,
      url: "http://example.com",
      username: "foo",
      password: "bar",
      external_url: "http://ext.example.com"
    )

    :ok
  end

  defp get_tiddler_returns_404(_) do
    TiddlyWikiMock
    |> stub(:get, fn _, _ -> {:error, %{status_code: 404}} end)

    :ok
  end

  @tag :tasks
  describe "/tasks" do
    test "send tasks to chat" do
      TiddlyWikiMock
      |> expect(:get_all, fn x, y ->
        assert "[tag[foo]]" == y

        assert %TiddlyWiki{
                 username: "foo",
                 password: "bar",
                 url: "http://example.com",
                 external_url: "http://ext.example.com"
               } == x

        {:ok, ["foo", "bar"] |> Enum.map(&%Tiddler{title: &1})}
      end)

      NadiaMock
      |> expect(:send_message, fn chat_id, text, _ ->
        assert 42 == chat_id

        assert "[foo](http://ext\\.example\\.com/foo)\n[bar](http://ext\\.example\\.com/bar)" ==
                 text

        {:ok, %{}}
      end)

      make_message(42, "/tasks", [
        %{
          "offset" => 0,
          "length" => 6,
          "type" => "bot_command"
        }
      ])
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end
  end

  @tag :links
  describe "create entry from links" do
    setup [:get_tiddler_returns_404]

    test "create entry in wiki from url" do
      ExtractInfoMock
      |> expect(:extract, fn url ->
        assert "http://example.com/new_entry" == url
        {:ok, %ExtractInfo{title: "New Entry"}}
      end)

      NadiaMock
      |> expect(:send_message, fn _, text, _ ->
        assert "http://ext\\.example\\.com\\#New%20Entry" == text

        {:ok, %{}}
      end)

      TiddlyWikiMock
      |> expect(:put, fn d, t ->
        assert %TiddlyWiki{
                 username: "foo",
                 password: "bar",
                 url: "http://example.com",
                 external_url: "http://ext.example.com"
               } == d

        assert t == %TiddlyWiki.Tiddler{
                 title: "New Entry",
                 tags: "external",
                 text: "http://example.com/new_entry",
                 fields: %{
                   chat_id: 42
                 }
               }

        {:ok, "http://ext.example.com#New%20Entry"}
      end)

      make_message(42, "http://example.com/new_entry")
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end

    @tag :multy
    test "create 1 or more entry for 1 or more url present in the messages" do
      update =
        make_message(23_334_928, "Cheap http://example.com finish http://example.com2", [
          %{"length" => 18, "offset" => 6, "type" => "url"},
          %{"length" => 19, "offset" => 32, "type" => "url"}
        ])

      TiddlyWikiMock
      |> expect(:put, 2, fn _, %{title: title} ->
        {:ok, title}
      end)

      NadiaMock
      |> expect(:send_message, 2, fn _, _, _ -> {:ok, %{}} end)

      ExtractInfoMock
      |> expect(:extract, 2, &{:error, &1})

      update
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end

    test "when extract info errors, url used as title" do
      expected_url = "http://ext.example.com#http%3A%2F%2Fexample.com%2Fnew_entry"

      ExtractInfoMock
      |> expect(:extract, &{:error, &1})

      NadiaMock
      |> expect(:send_message, fn _, text, _ ->
        assert "http://ext\\.example\\.com\\#http%3A%2F%2Fexample\\.com%2Fnew\\_entry" == text

        {:ok, %{}}
      end)

      TiddlyWikiMock
      |> expect(:put, fn _d, t ->
        assert t == %TiddlyWiki.Tiddler{
                 title: "http://example.com/new_entry",
                 tags: "external",
                 text: "http://example.com/new_entry",
                 fields: %{
                   chat_id: 42
                 }
               }

        {:ok, expected_url}
      end)

      make_message(42, "http://example.com/new_entry")
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end

    test "creates a unique title if a tiddler with the same title already exists" do
      ExtractInfoMock
      |> expect(:extract, fn _ ->
        {:ok, %ExtractInfo{title: "New Entry"}}
      end)

      NadiaMock
      |> expect(:send_message, fn _, _, _ -> {:ok, %{}} end)

      TiddlyWikiMock
      |> expect(:get, 2, fn _, title ->
        case title do
          "New Entry" -> {:ok, %TiddlyWiki.Tiddler{title: title}}
          _ -> {:error, %{status_code: 404}}
        end
      end)
      |> expect(:put, fn _, t ->
        assert t == %TiddlyWiki.Tiddler{
                 title: "New Entry - dup",
                 tags: "external",
                 text: "http://example.com/new_entry",
                 fields: %{
                   chat_id: 42
                 }
               }

        {:ok, "http://ext.example.com#New%20Entry%20-%20dup"}
      end)

      make_message(42, "http://example.com/new_entry")
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end

    test "if a duplicate already exists it does not override the original duplicate" do
      ExtractInfoMock
      |> expect(:extract, fn _ ->
        {:ok, %ExtractInfo{title: "New Entry"}}
      end)

      NadiaMock
      |> expect(:send_message, fn _, _, _ -> {:ok, %{}} end)

      TiddlyWikiMock
      |> expect(:get, 3, fn _, title ->
        case title do
          "New Entry" -> {:ok, %TiddlyWiki.Tiddler{title: title}}
          "New Entry - dup" -> {:ok, %TiddlyWiki.Tiddler{title: title}}
          _ -> {:error, %{status_code: 404}}
        end
      end)
      |> expect(:put, fn _, t ->
        assert t == %TiddlyWiki.Tiddler{
                 title: "New Entry - dup - dup",
                 tags: "external",
                 text: "http://example.com/new_entry",
                 fields: %{
                   chat_id: 42
                 }
               }

        {:ok, "http://ext.example.com#New%20Entry%20-%20dup%20-%20dup"}
      end)

      make_message(42, "http://example.com/new_entry")
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end

    @tag :tags
    test "non-url non-task message are saved as task for a tiddler by chat_id" do
      TiddlyWikiMock
      |> expect(:put, fn _, %{title: title} = t ->
        assert t.tags == "external"
        {:ok, title}
      end)

      NadiaMock
      |> expect(:send_message, 1, fn _, _, _ -> {:ok, %{}} end)

      ExtractInfoMock
      |> expect(:extract, 1, fn _ -> {:ok, %ExtractInfo{title: "New Entry"}} end)

      make_message(1001, "http://example.com/new_entry")
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)

      assert [
               %TiddlyWiki.Tiddler{
                 title: "New Entry",
                 tags: "external",
                 text: "http://example.com/new_entry",
                 fields: %{
                   chat_id: 1001
                 }
               }
             ] ==
               TiddlyWikiBot.Handler.get(1001)

      TiddlyWikiMock
      |> expect(:put, fn _, %{title: title} = t ->
        assert t.tags == "external adding [some tags] foo"
        {:ok, title}
      end)

      make_message(1001, "Adding [Some tagS] fOO", [])
      |> pop_in(["message", "entities"])
      |> elem(1)
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end
  end

  describe "/create" do
    setup [:get_tiddler_returns_404]

    test "create tiddler with title and text equal to the rest of the message" do
      NadiaMock
      |> expect(:send_message, fn _, text, _ ->
        assert "http://ext\\.example\\.com\\#some%20text%3A%20cool" == text

        {:ok, %{}}
      end)

      TiddlyWikiMock
      |> expect(:put, fn _, t ->
        assert t == %TiddlyWiki.Tiddler{
                 title: "some text: cool",
                 tags: "external",
                 text: "some text: cool",
                 fields: %{
                   chat_id: 42
                 }
               }

        {:ok, "http://ext.example.com#some%20text%3A%20cool"}
      end)

      make_message(42, "/create some text: cool")
      |> wrap_in_connection
      |> TiddlyWikiBot.Handler.call(nil)
    end
  end

  for {
        type,
        content,
        expected_file_name,
        expected_file_type,
        file_content
      } <- [
        {"document",
         %{
           "file_id" => "somefileid42",
           "file_name" => "some-document.pdf",
           "file_size" => 358_275,
           "file_unique_id" => "AgADSAsAAg7FGEk",
           "mime_type" => "application/pdf"
         }, "some-document.pdf", "application/pdf", <<37, 80, 68, 70, 45>> <> "foo"},
        {"photo",
         [
           %{
             "file_id" => "somefileid42",
             "file_name" => "some-image.png",
             "file_size" => 358_275,
             "file_unique_id" => "AgADSAsAAg7FGEk",
             "height" => 1280,
             "width" => 720
           },
           %{
             "file_id" => "somefileid42",
             "file_name" => "some-image.png",
             "file_size" => 358_275,
             "file_unique_id" => "AgADSAsAAg7FGEk",
             "height" => 320,
             "width" => 180
           }
         ], "some-image.png", "application/png", <<137, 80, 78, 71, 13, 10, 26, 10>> <> "foo"}
      ] do
    describe "#{type} upload" do
      setup [:get_tiddler_returns_404]

      @tag :upload
      test "create tiddler with the content of the file in base 64", %{bypass: bypass} do
        NadiaMock
        |> expect(:get_file, fn file_id ->
          assert "somefileid42" == file_id

          {:ok, %Nadia.Model.File{file_id: file_id}}
        end)
        |> expect(:get_file_link, fn _ ->
          {:ok, "http://localhost:#{bypass.port}/file/somefile42.pdf"}
        end)

        Bypass.expect(bypass, fn conn ->
          assert "/file/somefile42.pdf" == conn.request_path
          assert "GET" == conn.method

          conn
          |> Plug.Conn.resp(200, unquote(Macro.escape(file_content)))
        end)

        TiddlyWikiMock
        |> expect(:put, fn d, t ->
          assert %TiddlyWiki{
                   username: "foo",
                   password: "bar",
                   url: "http://example.com",
                   external_url: "http://ext.example.com"
                 } == d

          assert %TiddlyWiki.Tiddler{
                   title: unquote(Macro.escape(expected_file_name)),
                   type: unquote(Macro.escape(expected_file_type)),
                   tags: "external",
                   text: unquote(Macro.escape(Base.encode64(file_content))),
                   fields: %{
                     chat_id: 23_334_928
                   }
                 } = t

          {:ok, "http://ext.example.com#some-document.pdf"}
        end)

        NadiaMock
        |> expect(:send_message, fn _, text, _ ->
          assert "http://ext\\.example\\.com\\#some\\-document\\.pdf" == text

          {:ok, %{}}
        end)

        %{
          "message" => %{
            "chat" => %{
              "id" => 23_334_928
            }
          }
        }
        |> put_in(["message", unquote(type)], unquote(Macro.escape(content)))
        |> wrap_in_connection
        |> TiddlyWikiBot.Handler.call(nil)
      end
    end
  end

  describe "error handling" do
    @tag :errors
    test "it send a 200 response when the message is not handled" do
      conn =
        %{
          "message" => %{
            "chat" => %{
              "id" => 23_334_928
            },
            "unknow_type" => %{
              "file_id" => "somefileid42",
              "file_name" => "some-document.pdf",
              "file_size" => 358_275,
              "file_unique_id" => "AgADSAsAAg7FGEk",
              "mime_type" => "application/pdf"
            }
          }
        }
        |> wrap_in_connection
        |> TiddlyWikiBot.Handler.call(nil)

      assert 200 == conn.status
    end
  end
end
