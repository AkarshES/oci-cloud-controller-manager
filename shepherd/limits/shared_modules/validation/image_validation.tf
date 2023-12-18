variable "image_mapping_values" {

  validation {
    # Check if number of image regex matches present is equal to the number of kubernetes versions
    condition = (length(regexall("\\\"(([A-Za-z0-9]+(-[A-Za-z0-9]+)+)|([A-Za-z0-9]+))\\.[0-9]+-[A-Za-z0-9]+-[0-9]+@sha256:[A-Za-z0-9]+\\\"", var.image_mapping_values)) == (length(regexall("\\\"v\\d\\.[0-9]+\\\"", var.image_mapping_values)) + 1)) || (var.image_mapping_values == "true") || (var.image_mapping_values == "false") || (var.image_mapping_values == "{}") || (var.image_mapping_values == null)
    error_message = "Image Regex Validation Failed. Please make sure that all image urls passed are correct."
  }
}
