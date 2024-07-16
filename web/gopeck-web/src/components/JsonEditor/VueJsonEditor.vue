<template>
    <div
        className="svelte-jsoneditor-vue"
        ref="editor"></div>
</template>

<script>
import { JSONEditor } from 'vanilla-jsoneditor'

// https://codesandbox.io/p/sandbox/svelte-jsoneditor-vue-toln3w?
// JSONEditor properties as of version 0.3.60
const propNames = [
    'content',
    'mode',
    'mainMenuBar',
    'navigationBar',
    'statusBar',
    'readOnly',
    'indentation',
    'tabSize',
    'escapeControlCharacters',
    'escapeUnicodeCharacters',
    'validator',
    'onError',
    'onChange',
    'onChangeMode',
    'onClassName',
    'onRenderValue',
    'onRenderMenu',
    'queryLanguages',
    'queryLanguageId',
    'onChangeQueryLanguage',
    'onFocus',
    'onBlur',
]

function pickDefinedProps(object, propNames) {
    const props = {}
    for (const propName of propNames) {
        if (object[propName] !== undefined) {
            props[propName] = object[propName]
        }
    }
    return props
}

export default {
    name: 'VueJsonEditor',
    props: propNames,
    mounted() {
        this.editor = new JSONEditor({
            target: this.$refs['editor'],
            props: pickDefinedProps(this, propNames),
        })
        console.log('create editor', this.editor)
    },
    updated() {
        const props = pickDefinedProps(this, propNames)
        this.editor.updateProps(props)
    },
    beforeUnmount() {
        console.log('destroy editor')
        this.editor.destroy()
        this.editor = null
    },
}
</script>

<style scoped>
.svelte-jsoneditor-vue {
    display: flex;
    flex: 1;
}
</style>
