#! /bin/bash

# Merge output files
stats_directory="benchmark/stats"
rm "${stats_directory}" -rf
mkdir "${stats_directory}"

for file in benchmark/*/*/*/output.log; do
    program="${file#*/}"
    program="${program%%/*}"

    bench="${file#*/}"
    bench="${bench#*/}"
    bench="${bench%%/*}"

    config="${file#*/}"
    config="${config#*/}"
    config="${config#*/}"
    config="${config%%/*}"

    target_file_name="${program}_${bench}_${config}.log"
    cp "$file" "${stats_directory}/${target_file_name}"
done