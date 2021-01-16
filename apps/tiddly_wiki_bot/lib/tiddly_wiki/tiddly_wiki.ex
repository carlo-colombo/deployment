defmodule TiddlyWiki do
  defstruct [
    :url,
    :external_url,
    :username,
    :password
  ]

  @type t :: %TiddlyWiki{
          url: String.t(),
          external_url: String.t(),
          username: String.t(),
          password: String.t()
        }

  @callback get_all(dest :: TiddlyWiki.t(), filter :: String.t()) ::
              {:ok, list(TiddlyWiki.Tiddler.t())}
  @callback get(dest :: TiddlyWiki.t(), title :: String.t()) :: {:ok, TiddlyWiki.Tiddler.t()}
  @callback put(dest :: TiddlyWiki.t(), tiddler :: TiddlyWiki.Tiddler.t()) :: {:ok, String.t()}

  defp from_env do
    TiddlyWiki
    |> struct(Application.get_env(:tiddlywiki_bot, :wiki))
  end

  def put(%TiddlyWiki.Tiddler{} = tiddler) do
    from_env()
    |> implementation().put(tiddler)
  end

  def get(title) do
    from_env()
    |> implementation().get(title)
  end

  def get_all(filter) do
    from_env()
    |> implementation().get_all(filter)
  end

  def get_all do
    from_env()
    |> implementation().get_all()
  end

  def absolute_url(tiddler) do
    "#{from_env().external_url}/#{tiddler.title}"
  end

  defp implementation,
    do:
      Application.get_env(
        :tiddlywiki_bot,
        :tiddlywiki_client
      )
end
