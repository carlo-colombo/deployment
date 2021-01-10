use Mix.Config

config :tiddlywiki_bot,
  telegram_client: Nadia,
  tiddlywiki_client: TiddlyWiki.Client,
  extract_info_client: ExtractInfo.Client,
  tasks_filter: "[tag[chandler]!tag[done]]"

if Mix.env() == :test, do: import_config("#{Mix.env()}.exs")
