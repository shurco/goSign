/**
 * GoSign Embed SDK
 * Simple JavaScript SDK for embedding document signing
 * @version 1.0.0
 */

(function(window) {
  'use strict';

  class GoSignEmbed {
    constructor(options) {
      this.options = {
        slug: options.slug || '',
        container: options.container || document.body,
        width: options.width || '100%',
        height: options.height || '600px',
        baseURL: options.baseURL || window.location.origin,
        onReady: options.onReady || null,
        onOpened: options.onOpened || null,
        onFieldFilled: options.onFieldFilled || null,
        onCompleted: options.onCompleted || null,
        onDeclined: options.onDeclined || null,
        onError: options.onError || null,
      };

      this.iframe = null;
      this.isReady = false;

      this._init();
    }

    _init() {
      // Create iframe
      this.iframe = document.createElement('iframe');
      this.iframe.src = `${this.options.baseURL}/embed/${this.options.slug}`;
      this.iframe.style.width = this.options.width;
      this.iframe.style.height = this.options.height;
      this.iframe.style.border = 'none';
      this.iframe.allow = 'camera;microphone';

      // Get container element
      const container = typeof this.options.container === 'string'
        ? document.querySelector(this.options.container)
        : this.options.container;

      if (!container) {
        console.error('GoSignEmbed: Container not found');
        return;
      }

      // Add iframe to container
      container.appendChild(this.iframe);

      // Listen for messages from iframe
      window.addEventListener('message', this._handleMessage.bind(this));
    }

    _handleMessage(event) {
      // Check message source
      if (event.data && event.data.source === 'gosign-embed') {
        const { event: eventType, data } = event.data;

        switch (eventType) {
          case 'ready':
            this.isReady = true;
            if (this.options.onReady) {
              this.options.onReady(data);
            }
            break;

          case 'opened':
            if (this.options.onOpened) {
              this.options.onOpened(data);
            }
            break;

          case 'field_filled':
            if (this.options.onFieldFilled) {
              this.options.onFieldFilled(data);
            }
            break;

          case 'completed':
            if (this.options.onCompleted) {
              this.options.onCompleted(data);
            }
            break;

          case 'declined':
            if (this.options.onDeclined) {
              this.options.onDeclined(data);
            }
            break;

          case 'error':
            if (this.options.onError) {
              this.options.onError(data);
            }
            break;
        }
      }
    }

    // Public methods

    /**
     * Sends message to iframe
     */
    sendMessage(event, data) {
      if (!this.isReady) {
        console.warn('GoSignEmbed: Iframe not ready yet');
        return;
      }

      this.iframe.contentWindow.postMessage({
        source: 'gosign-parent',
        event,
        data
      }, '*');
    }

    /**
     * Closes/removes iframe
     */
    destroy() {
      if (this.iframe && this.iframe.parentNode) {
        this.iframe.parentNode.removeChild(this.iframe);
      }
      window.removeEventListener('message', this._handleMessage.bind(this));
      this.iframe = null;
      this.isReady = false;
    }

    /**
     * Changes iframe size
     */
    resize(width, height) {
      if (this.iframe) {
        if (width) this.iframe.style.width = width;
        if (height) this.iframe.style.height = height;
      }
    }
  }

  // Export SDK
  window.GoSignEmbed = GoSignEmbed;

  // Also create factory function for convenience
  window.createGoSignEmbed = function(options) {
    return new GoSignEmbed(options);
  };

})(window);

