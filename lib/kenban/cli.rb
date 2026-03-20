require "tempfile"

module Kenban
  class CLI
    COMMANDS = {
      "add"      => :cmd_add,
      "list"     => :cmd_list,
      "state"    => :cmd_state,
      "move"     => :cmd_move,
      "done"     => :cmd_done,
      "edit"     => :cmd_edit,
      "projects" => :cmd_projects,
      "help"     => :cmd_help,
    }.freeze

    def initialize(args, store: nil)
      @args = args
      @store = store || Store.new
    end

    def run
      command = @args.shift
      if command.nil? || command == "help" || command == "--help" || command == "-h"
        cmd_help
        return
      end

      handler = COMMANDS[command]
      unless handler
        $stderr.puts "Unknown command: #{command}"
        $stderr.puts "Run 'kenban help' for usage."
        exit 1
      end

      send(handler)
    end

    private

    # kenban add "[project] description"
    def cmd_add
      input = @args.join(" ")
      if input.empty?
        $stderr.puts "Usage: kenban add \"[project] description\""
        exit 1
      end

      begin
        task = Task.parse_add_input(input)
      rescue ArgumentError => e
        $stderr.puts "Error: #{e.message}"
        exit 1
      end

      @store.append_task(task)
      puts "Added: #{task}"
    end

    # kenban list [project]
    def cmd_list
      tasks = @store.read_tasks
      filter = @args.first

      tasks.each_with_index do |task, i|
        next if filter && !task.matches_project?(filter)
        puts "#{(i + 1).to_s.rjust(3)}. #{task}"
      end
    end

    # kenban state <state>
    def cmd_state
      state_name = @args.first
      unless state_name
        $stderr.puts "Usage: kenban state <#{Task::VALID_STATES.join('|')}>"
        exit 1
      end

      unless Task::VALID_STATES.include?(state_name.downcase)
        $stderr.puts "Unknown state: #{state_name}"
        $stderr.puts "Valid states: #{Task::VALID_STATES.join(', ')}"
        exit 1
      end

      tasks = @store.read_tasks
      tasks.each_with_index do |task, i|
        next unless task.matches_state?(state_name)
        puts "#{(i + 1).to_s.rjust(3)}. #{task}"
      end
    end

    # kenban move <number> <state>
    def cmd_move
      num_str, new_state = @args[0], @args[1]
      unless num_str && new_state
        $stderr.puts "Usage: kenban move <task_number> <state>"
        exit 1
      end

      index = parse_task_number(num_str)
      new_state = new_state.downcase
      unless Task::VALID_STATES.include?(new_state)
        $stderr.puts "Unknown state: #{new_state}"
        $stderr.puts "Valid states: #{Task::VALID_STATES.join(', ')}"
        exit 1
      end

      tasks = @store.read_tasks
      validate_index(index, tasks.size)

      task = tasks[index]
      old_state = task.state
      task.state = new_state
      @store.write_tasks(tasks)
      puts "Moved: [#{old_state}] -> [#{new_state}] #{task.to_s.sub(/\A\[\w+\]\s+/, '')}"
    end

    # kenban done <number>
    def cmd_done
      num_str = @args.first
      unless num_str
        $stderr.puts "Usage: kenban done <task_number>"
        exit 1
      end

      @args[1] = "done"
      cmd_move
    end

    # kenban edit <number>
    def cmd_edit
      num_str = @args.first
      unless num_str
        $stderr.puts "Usage: kenban edit <task_number>"
        exit 1
      end

      index = parse_task_number(num_str)
      tasks = @store.read_tasks
      validate_index(index, tasks.size)

      task = tasks[index]
      editor = ENV["EDITOR"] || "vi"

      tmp = Tempfile.new("kenban-edit")
      tmp.write(task.to_s)
      tmp.close

      system(editor, tmp.path)

      new_line = File.read(tmp.path).strip
      tmp.unlink

      if new_line.empty?
        $stderr.puts "Edit cancelled (empty line)."
        exit 1
      end

      begin
        new_task = Task.parse(new_line)
      rescue ArgumentError => e
        $stderr.puts "Invalid task format: #{e.message}"
        $stderr.puts "Original task preserved."
        exit 1
      end

      tasks[index] = new_task
      @store.write_tasks(tasks)
      puts "Updated: #{new_task}"
    end

    # kenban projects
    def cmd_projects
      tasks = @store.read_tasks
      projects = tasks.map(&:project).uniq.sort_by(&:downcase)
      projects.each { |p| puts p }
    end

    def cmd_help
      puts <<~HELP
        kenban - plain-text kanban for solo developers

        Usage:
          kenban add "[project] description"   Add a new task (state: todo)
          kenban list [project]                List tasks, optionally filtered by project
          kenban state <state>                 List tasks filtered by state
          kenban move <number> <state>         Move a task to a new state
          kenban done <number>                 Shortcut: move task to done
          kenban edit <number>                 Edit a task in $EDITOR
          kenban projects                      List all project names

        States: #{Task::VALID_STATES.join(', ')}

        Task format:
          [state] [project] description @tag +label

        File location:
          ./tasks.txt (if present) or ~/.kenban/tasks.txt
      HELP
    end

    def parse_task_number(str)
      num = Integer(str, exception: false)
      unless num && num >= 1
        $stderr.puts "Invalid task number: #{str}"
        exit 1
      end
      num - 1 # convert to 0-based index
    end

    def validate_index(index, size)
      unless index >= 0 && index < size
        $stderr.puts "Task number out of range. You have #{size} task(s)."
        exit 1
      end
    end
  end
end
