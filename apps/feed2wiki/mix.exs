defmodule Feed2wiki.MixProject do
  use Mix.Project

  def project do
    [
      app: :feed2wiki,
      version: "0.1.0",
      elixir: "~> 1.11",
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
      extra_applications: [:logger, :xmerl],
      mod: {Feed2wiki.Application, [Mix.env()]}
    ]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    [
      {:libcluster, "~> 3.2"},
      {:exsync, "~> 0.2", only: :dev},
      {:feeder_ex, "~> 1.1.0"},
      {:mojito, "~> 0.7.1"},
      {:jason, "~> 1.1"},
      {:mix_test_watch, "~> 1.0", only: :dev, runtime: false},
      {:common, in_umbrella: true}
    ]
  end
end
