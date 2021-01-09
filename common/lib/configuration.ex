defmodule Common.Configuration do

  def libcluster do
    {:ok, hostname} = :inet.gethostname()

    [topologies: [
        connection: [
          strategy: Elixir.Cluster.Strategy.Epmd,
          config: [
            hosts:
            ["dashboard", "tiddlywiki_bot", "feed2wiki"]
            |> Enum.map(&Enum.join([&1, hostname], "@"))
            |> Enum.map(&String.to_atom/1)
          ]
        ]
      ]]
  end
end
