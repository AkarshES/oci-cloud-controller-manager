import importlib.util

# read and load release schedule
release_schedule_path = importlib.util.spec_from_file_location("release_schedules", "templates/shared_modules/release_templates/release_schedules.py")
release_schedule_module = importlib.util.module_from_spec(release_schedule_path)
release_schedule_path.loader.exec_module(release_schedule_module)

# Read and load excluded locations 
excluded_locations_path = importlib.util.spec_from_file_location("excluded_locations", "templates/shared_modules/release_templates/excluded_locations.py")
excluded_locations_module = importlib.util.module_from_spec(excluded_locations_path)
excluded_locations_path.loader.exec_module(excluded_locations_module)

# Assigning the release schedule and excluded locations
oke_common_release_schedule = release_schedule_module.oke_common_release_schedule_non_cell_based

for bundle in oke_common_release_schedule:
    for group in bundle["groups"]:
        group["et_sequence"] = "true"

excluded_locations = excluded_locations_module.excluded_locations
