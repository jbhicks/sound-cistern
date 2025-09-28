/**
 * Quill Rich Text Editor Initialization (v2.0.3)
 * Initializes Quill editor for content fields in admin forms
 */

// Initialize Quill editor when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Check if we're on a page with a Quill editor
    const editorContainer = document.getElementById('quill-editor');
    const contentTextarea = document.getElementById('content');
    
    if (editorContainer && contentTextarea) {
        // Configure Quill toolbar for v2.x
        const toolbarOptions = [
            [{ 'header': [2, 3, false] }],
            ['bold', 'italic', 'underline', 'strike'],
            [{ 'list': 'ordered'}, { 'list': 'bullet' }],
            ['blockquote', 'code-block'],
            ['link'],
            [{ 'align': [] }],
            ['clean']
        ];

        // Initialize Quill with v2.x API
        const quill = new Quill('#quill-editor', {
            theme: 'snow',
            modules: {
                toolbar: toolbarOptions
            },
            placeholder: 'Write your post content here...'
        });

        // Set initial content from textarea
        const initialContent = contentTextarea.value;
        if (initialContent && initialContent.trim()) {
            // Use innerHTML for compatibility with both v1.x and v2.x
            quill.root.innerHTML = initialContent;
        }

        // Update hidden textarea when Quill content changes
        quill.on('text-change', function() {
            // Use getSemanticHTML() if available (v2.x), fallback to innerHTML
            if (typeof quill.getSemanticHTML === 'function') {
                contentTextarea.value = quill.getSemanticHTML();
            } else {
                contentTextarea.value = quill.root.innerHTML;
            }
        });

        // Ensure content is synced before form submission
        const form = editorContainer.closest('form');
        if (form) {
            form.addEventListener('submit', function() {
                if (typeof quill.getSemanticHTML === 'function') {
                    contentTextarea.value = quill.getSemanticHTML();
                } else {
                    contentTextarea.value = quill.root.innerHTML;
                }
            });
        }

        // Style the editor container to match Pico.css theme
        updateQuillTheme();
        
        // Listen for theme changes
        const observer = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                if (mutation.type === 'attributes' && mutation.attributeName === 'data-theme') {
                    updateQuillTheme();
                }
            });
        });
        observer.observe(document.documentElement, { attributes: true });
    }
    
    function updateQuillTheme() {
        const editor = document.querySelector('.ql-editor');
        const toolbar = document.querySelector('.ql-toolbar');
        
        if (editor && toolbar) {
            // Apply Pico.css variables
            editor.style.fontFamily = 'var(--pico-font-family)';
            editor.style.fontSize = 'var(--pico-font-size)';
            editor.style.lineHeight = 'var(--pico-line-height)';
            editor.style.color = 'var(--pico-color)';
            editor.style.backgroundColor = 'var(--pico-background-color)';
            
            toolbar.style.backgroundColor = 'var(--pico-card-background-color)';
            toolbar.style.borderColor = 'var(--pico-border-color)';
        }
    }
});
