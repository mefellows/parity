#!/bin/bash -e

createuser --createdb --login -w ihi_pa_dev
psql -c "alter role \"ihi_pa_dev\" password 'password'" postgres
createuser --createdb --login -w ihi_pa_test
psql -c "alter role \"ihi_pa_test\" password 'password'" postgres
