defmodule TiddlyWiki.Tiddler do
  @derive Jason.Encoder
  defstruct [
    :title,
    :tags,
    :modified,
    :created,
    :text,
    :creator,
    :modifier,
    :fields,
    :type
  ]

  @type t :: %TiddlyWiki.Tiddler{
          title: String.t(),
          tags: String.t(),
          modified: String.t(),
          created: String.t(),
          type: String.t(),
          text: String.t()
        }
end
