defmodule TiddlywikiBotTest do
  use ExUnit.Case
  doctest TiddlywikiBot

  test "greets the world" do
    assert TiddlywikiBot.hello() == :world
  end
end
