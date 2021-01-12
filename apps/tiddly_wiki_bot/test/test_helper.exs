ExUnit.start()

Mox.defmock(TiddlyWikiMock, for: TiddlyWiki)
Mox.defmock(NadiaMock, for: Nadia.Behaviour)
Mox.defmock(ExtractInfoMock, for: ExtractInfo)

defmodule TestHelper do
  def make_message(chat_id, text) do
    make_message(chat_id, text, [
      %{"length" => String.length(text), "offset" => 0, "type" => "url"}
    ])
  end

  def make_message(chat_id, text, entities),
    do: %{
      "message" => %{
        "chat" => %{
          "id" => chat_id
        },
        "text" => text,
        "entities" => entities
      }
    }

  def wrap_in_connection(message) do
    Plug.Test.conn(:post, "/", "")
    |> Map.put(:body_params, message)
  end
end
