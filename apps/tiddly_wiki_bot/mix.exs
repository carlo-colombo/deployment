defmodule TiddlywikiBot.MixProject do
  use Mix.Project

  def project do
    [
      app: :tiddly_wiki_bot,
      version: "0.1.0",
      elixir: "~> 1.11.0-rc.0",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      build_path: "../../_build",
      config_path: "../../config/config.exs",
      deps_path: "../../deps",
      lockfile: "../../mix.lock"
    ]
  end

  # Run "mix help compile.app" to learn about applications.
  def application do
    [
      extra_applications: [:logger],
      mod: {TiddlyWikiBot.Application, [Mix.env()]}
    ]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    [
      {:plug_cowboy, "~> 2.0"},
      {:jason, "~> 1.1"},
      {:nadia, "~> 0.7.0"},
      {:mojito, "~> 0.7.1"},
      {:libcluster, "~> 3.2"},
      {:bypass, "2.1.0"},
      {:mox, "~> 1.0", only: :test},
      {:exsync, "~> 0.2", only: :dev},
      {:mix_test_watch, "~> 1.0", only: :dev, runtime: false},
      {:credo, "~> 1.6.6", only: [:dev, :test], runtime: false},
      {:common, in_umbrella: true}
    ]
  end
end
