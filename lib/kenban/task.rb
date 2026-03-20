module Kenban
  class Task
    VALID_STATES = %w[todo doing blocked done].freeze
    LINE_PATTERN = /\A\[(\w+)\]\s+\[([^\]]+)\]\s+(.+)\z/

    attr_accessor :state, :project, :description

    def initialize(state:, project:, description:)
      @state = state
      @project = project
      @description = description
    end

    # Parse a single line into a Task. Returns nil if blank.
    # Raises ArgumentError if malformed.
    def self.parse(line)
      stripped = line.to_s.strip
      return nil if stripped.empty?

      match = LINE_PATTERN.match(stripped)
      unless match
        raise ArgumentError, "Malformed task line: #{stripped.inspect}"
      end

      state = match[1].downcase
      unless VALID_STATES.include?(state)
        raise ArgumentError, "Unknown state [#{match[1]}]. Valid states: #{VALID_STATES.join(', ')}"
      end

      project = match[2]
      description = match[3].strip

      if description.empty?
        raise ArgumentError, "Task description cannot be empty"
      end

      new(state: state, project: project, description: description)
    end

    # Parse project and description from add input like "[project] description"
    def self.parse_add_input(input)
      input = input.to_s.strip
      match = /\A\[([^\]]+)\]\s+(.+)\z/.match(input)
      unless match
        raise ArgumentError,
          "Expected format: [project] description\n" \
          "Example: kenban add \"[myproject] Fix the login bug\""
      end

      project = match[1]
      description = match[2].strip

      if description.empty?
        raise ArgumentError, "Task description cannot be empty"
      end

      new(state: "todo", project: project, description: description)
    end

    def to_s
      "[#{state}] [#{project}] #{description}"
    end

    def matches_project?(name)
      project.downcase == name.downcase
    end

    def matches_state?(name)
      state.downcase == name.downcase
    end
  end
end
