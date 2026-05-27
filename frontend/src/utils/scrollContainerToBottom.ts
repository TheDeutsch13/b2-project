export function scrollContainerToBottom(
  container: HTMLElement | null,
  smooth = false
) {
  if (!container) {
    return;
  }

  container.scrollTo({
    top: container.scrollHeight,
    behavior: smooth ? "smooth" : "auto",
  });
}
