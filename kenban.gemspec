Gem::Specification.new do |s|
  s.name        = "kenban"
  s.version     = "0.1.0"
  s.summary     = "Plain-text personal kanban for solo developers"
  s.description = "A CLI task manager that stores tasks in a single plain-text file. Trello for one developer, Unix-style."
  s.authors     = ["kenban contributors"]
  s.homepage    = "https://github.com/kenban/kenban"
  s.license     = "MIT"

  s.required_ruby_version = ">= 2.7.0"

  s.files       = Dir["lib/**/*.rb", "bin/*", "README.md"]
  s.executables = ["kenban", "kb"]
  s.require_paths = ["lib"]

  s.add_development_dependency "minitest", "~> 5.0"
end
