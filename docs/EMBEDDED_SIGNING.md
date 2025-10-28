# Embedded Signing - Documentation

## Introduction

GoSign supports embedding the document signing interface into your application via iframe. This allows users to sign documents without navigating to a separate website.

## Quick Start

### 1. SDK Integration

```html
<!DOCTYPE html>
<html>
<head>
    <title>Embedded Signing Example</title>
</head>
<body>
    <!-- Container for iframe -->
    <div id="signing-container"></div>

    <!-- Include SDK -->
    <script src="https://yourdomain.com/gosign-embed.js"></script>

    <script>
        // Initialize embedded signing
        const signer = new GoSignEmbed({
            slug: 'abc123xyz', // Signer slug
            container: '#signing-container',
            width: '100%',
            height: '800px',

            // Event callbacks
            onReady: function(data) {
                console.log('Signing interface ready:', data);
            },

            onCompleted: function(data) {
                console.log('Document completed:', data);
                alert('Document signed successfully!');
            },

            onDeclined: function(data) {
                console.log('Document declined:', data);
                alert('Signing declined');
            },

            onError: function(error) {
                console.error('Error:', error);
            }
        });
    </script>
</body>
</html>
```

### 2. React Example

```jsx
import React, { useEffect, useRef } from 'react';

function SigningComponent({ slug }) {
  const containerRef = useRef(null);
  const embedRef = useRef(null);

  useEffect(() => {
    // Create GoSignEmbed instance
    embedRef.current = new window.GoSignEmbed({
      slug: slug,
      container: containerRef.current,
      width: '100%',
      height: '800px',

      onReady: (data) => {
        console.log('Ready:', data);
      },

      onCompleted: (data) => {
        console.log('Completed:', data);
        // Update your app state
      },

      onDeclined: (data) => {
        console.log('Declined:', data);
      }
    });

    // Cleanup on unmount
    return () => {
      if (embedRef.current) {
        embedRef.current.destroy();
      }
    };
  }, [slug]);

  return <div ref={containerRef} />;
}

export default SigningComponent;
```

### 3. Vue 3 Example

```vue
<template>
  <div ref="containerRef"></div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const props = defineProps({
  slug: {
    type: String,
    required: true
  }
});

const containerRef = ref(null);
let embedInstance = null;

onMounted(() => {
  embedInstance = new window.GoSignEmbed({
    slug: props.slug,
    container: containerRef.value,
    width: '100%',
    height: '800px',

    onReady: (data) => {
      console.log('Ready:', data);
    },

    onCompleted: (data) => {
      console.log('Completed:', data);
    },

    onDeclined: (data) => {
      console.log('Declined:', data);
    }
  });
});

onUnmounted(() => {
  if (embedInstance) {
    embedInstance.destroy();
  }
});
</script>
```

## API Reference

### Constructor Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `slug` | string | Yes | - | Signer slug from URL `/s/:slug` |
| `container` | string\|Element | No | document.body | CSS selector or DOM element |
| `width` | string | No | '100%' | Iframe width |
| `height` | string | No | '600px' | Iframe height |
| `baseURL` | string | No | window.location.origin | GoSign server base URL |
| `onReady` | function | No | null | Callback when ready |
| `onOpened` | function | No | null | Callback when document opened |
| `onFieldFilled` | function | No | null | Callback when field filled |
| `onCompleted` | function | No | null | Callback when completed |
| `onDeclined` | function | No | null | Callback when declined |
| `onError` | function | No | null | Callback on error |

### Methods

#### `sendMessage(event, data)`
Sends custom message to iframe.

```javascript
signer.sendMessage('custom_event', { foo: 'bar' });
```

#### `destroy()`
Removes iframe and cleans up event listeners.

```javascript
signer.destroy();
```

#### `resize(width, height)`
Changes iframe size.

```javascript
signer.resize('100%', '1000px');
```

## Events

### ready
Called when iframe is fully loaded and ready to work.

```javascript
onReady: function(data) {
  console.log('Slug:', data.slug);
}
```

### opened
Called when signer opens document for the first time.

```javascript
onOpened: function(data) {
  console.log('Document opened at:', data.timestamp);
}
```

### field_filled
Called when each field is filled.

```javascript
onFieldFilled: function(data) {
  console.log('Field filled:', data.field_id, data.value);
}
```

### completed
Called when document is successfully signed.

```javascript
onCompleted: function(data) {
  console.log('Submission ID:', data.submission_id);
  console.log('Completed at:', data.completed_at);

  // Show success message
  showSuccessNotification();
}
```

### declined
Called when signer declines the document.

```javascript
onDeclined: function(data) {
  console.log('Declined reason:', data.reason);
}
```

### error
Called on any error.

```javascript
onError: function(error) {
  console.error('Error:', error.message);
  showErrorNotification(error.message);
}
```

## Security

### CORS
Make sure your site is added to CORS whitelist on GoSign server.

### Content Security Policy
Add GoSign domain to CSP policy:

```html
<meta http-equiv="Content-Security-Policy"
      content="frame-src https://yourdomain.com">
```

### X-Frame-Options
GoSign automatically sets correct headers for embedding.

## Styling

You can style the iframe container:

```css
#signing-container {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  overflow: hidden;
}
```

## Use Case Examples

### 1. Embedding in Admin Panel

```javascript
const signer = new GoSignEmbed({
  slug: userSlug,
  container: '#admin-signing-panel',
  onCompleted: function(data) {
    // Update status in admin panel
    updateUserStatus(userId, 'signed');
    // Close modal
    closeModal();
  }
});
```

### 2. Multi-step Form

```javascript
// Step 1: Data filling
// Step 2: Embedded signing
// Step 3: Confirmation

function showSigningStep() {
  const signer = new GoSignEmbed({
    slug: generatedSlug,
    container: '#step-2-container',
    onCompleted: function(data) {
      // Proceed to next step
      showStep3();
    }
  });
}
```

### 3. Mobile Responsive

```javascript
function createResponsiveSigner() {
  const isMobile = window.innerWidth < 768;

  return new GoSignEmbed({
    slug: slug,
    container: '#signing-container',
    width: '100%',
    height: isMobile ? '100vh' : '800px'
  });
}
```

## FAQ

**Q: Can the appearance be customized?**
A: Yes, you can style the container via CSS. For deep customization, use the API to create your own UI.

**Q: Does it work on mobile devices?**
A: Yes, iframe is fully responsive and supports touch events.

**Q: Is authentication required?**
A: No, embedded signing uses public slugs. Each slug is unique and one-time use.

**Q: Can progress be tracked?**
A: Yes, use the `onFieldFilled` event to track each field being filled.

## Support

For questions and support: support@yourdomain.com

