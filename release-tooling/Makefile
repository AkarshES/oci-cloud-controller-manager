add-gitmodule:
	@echo "Adding oke-common-schedule to templates/shared_modules.."
	git submodule update --init --recursive
update-gitmodule:
	@echo "Discarding all local changes to release schedule"
	cd templates/shared_modules && git restore release_templates && cd ../../
	@echo "Updating submodule to latest commit on the master branch..."
	git submodule update --remote
