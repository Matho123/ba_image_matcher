import cv2


def calculate_rotation_invariant_orientation(image_path):
    # Load the image
    img = cv2.imread(image_path, cv2.IMREAD_GRAYSCALE)

    # Initialize the SIFT detector
    sift = cv2.SIFT_create()

    # Detect keypoints and compute descriptors
    keypoints, descriptors = sift.detectAndCompute(img, None)

    # Calculate the orientation from the keypoints
    total_orientation = 0.0

    for keypoint in keypoints:
        total_orientation += keypoint.angle

    # Calculate the average orientation
    average_orientation = total_orientation / len(keypoints)

    return average_orientation

image_path = "../images/variations/rotated/1014535523-rotated-45.png"  # Replace with the path to your image
orientation = calculate_rotation_invariant_orientation(image_path)
print(f"Rotation-Invariant Orientation: {orientation} degrees\n")
