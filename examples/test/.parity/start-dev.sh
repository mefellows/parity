#!/bin/bash

echo "Populating database and starting Rails..."
rake db:migrate:reset
rake db:seed
rake db:populate:random
bundle exec rails s -p 3000 -b '0.0.0.0'
