import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import InterpretationLevelSwitch from './InterpretationLevelSwitch.vue'

describe('InterpretationLevelSwitch', () => {
  it('renders the three levels and emits a model update', async () => {
    const wrapper = mount(InterpretationLevelSwitch, { props: { modelValue: 'basic' } })

    expect(wrapper.get('[data-level="basic"]').attributes('aria-checked')).toBe('true')
    expect(wrapper.text()).toContain('普通')
    expect(wrapper.text()).toContain('进阶')
    expect(wrapper.text()).toContain('专业')

    await wrapper.get('[data-level="professional"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['professional'])
  })
})
