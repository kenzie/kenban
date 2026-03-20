require "minitest/autorun"
$LOAD_PATH.unshift(File.expand_path("../lib", __dir__))
require "kenban"

class TestTask < Minitest::Test
  def test_parse_valid_line
    task = Kenban::Task.parse("[todo] [goaliebook] Stripe onboarding")
    assert_equal "todo", task.state
    assert_equal "goaliebook", task.project
    assert_equal "Stripe onboarding", task.description
  end

  def test_parse_with_tags
    task = Kenban::Task.parse("[doing] [kioskbook] Fix bug @high +payments")
    assert_equal "doing", task.state
    assert_equal "kioskbook", task.project
    assert_equal "Fix bug @high +payments", task.description
  end

  def test_parse_blank_line_returns_nil
    assert_nil Kenban::Task.parse("")
    assert_nil Kenban::Task.parse("   ")
  end

  def test_parse_malformed_line_raises
    assert_raises(ArgumentError) { Kenban::Task.parse("just some text") }
  end

  def test_parse_invalid_state_raises
    assert_raises(ArgumentError) { Kenban::Task.parse("[nope] [proj] desc") }
  end

  def test_parse_empty_description_raises
    assert_raises(ArgumentError) { Kenban::Task.parse("[todo] [proj]   ") }
  end

  def test_to_s_round_trips
    line = "[blocked] [teambook] Budget export waiting on schema"
    task = Kenban::Task.parse(line)
    assert_equal line, task.to_s
  end

  def test_parse_case_insensitive_state
    task = Kenban::Task.parse("[TODO] [proj] Something")
    assert_equal "todo", task.state
  end

  def test_parse_add_input
    task = Kenban::Task.parse_add_input("[goaliebook] Stripe onboarding")
    assert_equal "todo", task.state
    assert_equal "goaliebook", task.project
    assert_equal "Stripe onboarding", task.description
  end

  def test_parse_add_input_missing_project_raises
    assert_raises(ArgumentError) { Kenban::Task.parse_add_input("just text") }
  end

  def test_matches_project
    task = Kenban::Task.parse("[todo] [GoalieBook] Something")
    assert task.matches_project?("goaliebook")
    assert task.matches_project?("GoalieBook")
    refute task.matches_project?("other")
  end

  def test_matches_state
    task = Kenban::Task.parse("[doing] [proj] Something")
    assert task.matches_state?("doing")
    assert task.matches_state?("DOING")
    refute task.matches_state?("todo")
  end
end
