(function () {
  'use strict';
  const workflow = document.querySelector('.case-workflow');
  if (!workflow) return;
  let submitting = false;
  let observing = false;
  const feedback = workflow.querySelector('[data-workflow-feedback]');
  const quiz = workflow.querySelector('#quiz-surface');
  const submittedKey = quiz && 'chr-workflow-answer:' + quiz.dataset.quizSession;
  try {
    const previous = submittedKey && sessionStorage.getItem(submittedKey);
    if (previous && quiz.dataset.quizQuestion && previous !== quiz.dataset.quizQuestion) {
      const key = 'chr-quiz-drafts:' + quiz.dataset.quizContext + ':' + quiz.dataset.quizSession;
      const drafts = JSON.parse(sessionStorage.getItem(key) || '{}');
      delete drafts[previous];
      sessionStorage.setItem(key, JSON.stringify(drafts));
      sessionStorage.removeItem(submittedKey);
      // The shared quiz restores the remaining drafts on DOMContentLoaded.
    }
  } catch (_) { /* The shared quiz surface reports unavailable draft storage. */ }

  workflow.addEventListener('submit', function (event) {
    if (submitting) { event.preventDefault(); return; }
    submitting = true;
    if (quiz) {
      quiz.dataset.workflowSubmitting = 'true';
      if (event.target.elements.step && event.target.elements.step.value === 'answer') {
        try { sessionStorage.setItem(submittedKey, quiz.dataset.quizQuestion); } catch (_) {}
      }
    }
    if (feedback) feedback.textContent = 'Saving…';
    workflow.setAttribute('aria-busy', 'true');
  });

  async function observe() {
    if (document.visibilityState !== 'visible' || submitting || observing) return;
    observing = true;
    try {
      const response = await fetch(workflow.dataset.workflowUrl, {cache: 'no-store', signal: AbortSignal.timeout(8000), headers: {'Accept': 'application/json'}});
      if (!response.ok) throw new Error('refresh failed');
      const state = await response.json();
      if (state.state === workflow.dataset.workflowState) {
        if (feedback && feedback.dataset.unavailable) { feedback.textContent = ''; delete feedback.dataset.unavailable; }
        return;
      }
      const dirty = Array.from(workflow.querySelectorAll('textarea, input:not([type=hidden])')).some(function (input) { return input.value !== input.defaultValue; });
      if (dirty) {
        if (feedback && !feedback.querySelector('a')) {
          feedback.textContent = 'This workflow has changed. Your unsent text is still here. ';
          const link = document.createElement('a');
          link.href = location.href; link.textContent = 'Reload when ready';
          feedback.append(link);
        }
        return;
      }
      location.assign(state.url || location.href);
    } catch (_) {
      if (feedback) { feedback.textContent = 'Updates are temporarily unavailable. Your progress is saved.'; feedback.dataset.unavailable = 'true'; }
    } finally {
      observing = false;
    }
  }
  // Read-only observations follow host work; they never start or retry an operation.
  if (workflow.dataset.workflowWatch === 'true') setInterval(observe, 2500);
})();
