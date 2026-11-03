import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.tsx'

const fetchMock = vi.fn<typeof fetch>()

beforeEach(() => {
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  fetchMock.mockReset()
  vi.unstubAllGlobals()
})

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function respondWithResult(result: number) {
  fetchMock.mockImplementation(() => Promise.resolve(jsonResponse(200, { operation: 'add', result })))
}

function sentRequest(call = 0): { url: string; body: unknown } {
  const [url, init] = fetchMock.mock.calls[call]
  return { url: String(url), body: JSON.parse(String(init?.body)) }
}

function setup() {
  const user = userEvent.setup()
  render(<App />)
  return user
}

// Labels change per operation (see the label test), so tests reach the inputs by position.
const firstInput = () => screen.getAllByRole('textbox')[0]
const secondInput = (): HTMLElement | undefined => screen.getAllByRole('textbox')[1]
const calculateButton = () => screen.getByRole('button', { name: /calculat/i })
const result = () => screen.getByRole('status', { name: 'Result' })

async function fillOperands(user: ReturnType<typeof userEvent.setup>, a: string, b?: string) {
  await user.clear(firstInput())
  await user.type(firstInput(), a)
  const second = secondInput()
  if (b !== undefined && second) {
    await user.clear(second)
    await user.type(second, b)
  }
}

describe('App', () => {
  it('renders the heading, an empty result and an empty history', () => {
    setup()

    expect(screen.getByRole('heading', { name: 'Seezle Calculator' })).toBeInTheDocument()
    expect(result()).toHaveTextContent('0')
    expect(screen.getByText('No calculations yet.')).toBeInTheDocument()
    expect(calculateButton()).toBeDisabled()
  })

  it('sends the operands to the API and shows the returned result', async () => {
    respondWithResult(36)
    const user = setup()

    await user.click(screen.getByRole('radio', { name: 'Multiply' }))
    await fillOperands(user, '12', '3')
    await user.click(calculateButton())

    await waitFor(() => expect(result()).toHaveTextContent('36'))
    expect(result()).toHaveAccessibleDescription('12 × 3')
    expect(sentRequest()).toEqual({ url: '/api/v1/multiply', body: { a: 12, b: 3 } })
  })

  describe('validation', () => {
    it.each(['1.2.3', '--1', '1e', 'abc', '1,5'])('shows an inline message and disables Calculate for %j', async (raw) => {
      const user = setup()

      await fillOperands(user, raw, '2')

      expect(firstInput()).toHaveAttribute('aria-invalid', 'true')
      expect(firstInput()).toHaveAccessibleDescription('Enter a valid number, like -1.5 or 2e3.')
      expect(calculateButton()).toBeDisabled()
    })

    it('flags numbers too large to represent', async () => {
      const user = setup()

      await fillOperands(user, '1', '1e999')

      expect(secondInput()).toHaveAccessibleDescription('That number is too large.')
      expect(calculateButton()).toBeDisabled()
    })

    it('accepts sign, decimal point and scientific notation', async () => {
      const user = setup()

      await fillOperands(user, '-1.5e-3', '+.5')

      expect(screen.queryByText(/Enter a valid number/)).not.toBeInTheDocument()
      expect(firstInput()).not.toHaveAttribute('aria-invalid')
      expect(calculateButton()).toBeEnabled()
    })

    it('keeps Calculate disabled without a message while a field is empty', async () => {
      const user = setup()

      await fillOperands(user, '4')

      expect(secondInput()).not.toHaveAttribute('aria-invalid')
      expect(screen.queryByText('Enter a number.')).not.toBeInTheDocument()
      expect(calculateButton()).toBeDisabled()
    })
  })

  it('announces API errors in an alert', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(422, { error: { code: 'division_by_zero', message: 'division by zero' } }),
    )
    const user = setup()

    await user.click(screen.getByRole('radio', { name: 'Divide' }))
    await fillOperands(user, '1', '0')
    await user.click(calculateButton())

    expect(await screen.findByRole('alert')).toHaveTextContent("You can't divide by zero.")
    expect(result()).toHaveTextContent('0')
  })

  it('hides the second input for sqrt and sends only "a"', async () => {
    respondWithResult(9)
    const user = setup()

    await user.click(screen.getByRole('radio', { name: 'Square root' }))

    expect(secondInput()).toBeUndefined()
    await fillOperands(user, '81')
    await user.click(calculateButton())

    await waitFor(() => expect(result()).toHaveAccessibleDescription('√81'))
    expect(sentRequest()).toEqual({ url: '/api/v1/sqrt', body: { a: 81 } })

    await user.click(screen.getByRole('radio', { name: 'Add' }))
    expect(secondInput()).toBeInTheDocument()
  })

  it.each([
    ['Add', 'First number', 'Second number'],
    ['Subtract', 'First number', 'Second number'],
    ['Multiply', 'First number', 'Second number'],
    ['Divide', 'Dividend', 'Divisor'],
    ['Power', 'Base', 'Exponent'],
    ['Percentage', 'Percent', 'Of'],
  ])('labels the inputs for %s as %j and %j', async (operation, labelA, labelB) => {
    const user = setup()

    await user.click(screen.getByRole('radio', { name: operation }))

    expect(screen.getAllByRole('textbox')).toEqual([screen.getByLabelText(labelA), screen.getByLabelText(labelB)])
  })

  it('labels the single sqrt input "Number"', async () => {
    const user = setup()

    await user.click(screen.getByRole('radio', { name: 'Square root' }))

    expect(screen.getAllByRole('textbox')).toEqual([screen.getByLabelText('Number')])
  })

  it('shows a "0" placeholder in empty inputs', () => {
    setup()

    for (const input of screen.getAllByRole('textbox')) {
      expect(input).toHaveAttribute('placeholder', '0')
    }
  })

  it('shows the expression above the result, using the operands that were sent', async () => {
    respondWithResult(24600)
    const user = setup()

    expect(result()).not.toHaveAccessibleDescription()
    await user.click(screen.getByRole('radio', { name: 'Multiply' }))
    await fillOperands(user, '200', '123')
    await user.click(calculateButton())
    await waitFor(() => expect(result()).toHaveTextContent('24600'))
    await fillOperands(user, '7', '8')

    expect(result()).toHaveAccessibleDescription('200 × 123')
  })

  it('keeps the current operation when its segment is clicked again', async () => {
    const user = setup()

    await user.click(screen.getByRole('radio', { name: 'Add' }))

    expect(screen.getByRole('radio', { name: 'Add' })).toBeChecked()
  })

  it('disables Calculate while a request is pending and ignores further submits', async () => {
    let respond!: (response: Response) => void
    fetchMock.mockReturnValue(new Promise((resolve) => (respond = resolve)))
    const user = setup()

    await fillOperands(user, '2', '3')
    await user.click(calculateButton())

    expect(calculateButton()).toHaveAccessibleName('Calculating…')
    expect(calculateButton()).toBeDisabled()
    expect(calculateButton()).toHaveAttribute('aria-busy', 'true')
    await user.click(calculateButton())
    await user.type(firstInput(), '{Enter}')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    respond(jsonResponse(200, { operation: 'add', result: 5 }))

    await waitFor(() => expect(calculateButton()).toHaveAccessibleName('Calculate'))
    expect(result()).toHaveTextContent('5')
  })

  it('calculates on Enter and clears on Escape', async () => {
    respondWithResult(5)
    const user = setup()

    await user.type(firstInput(), '2')
    await user.type(secondInput()!, '3{Enter}')

    await waitFor(() => expect(result()).toHaveTextContent('5'))
    expect(result()).toHaveAccessibleDescription('2 + 3')

    await user.keyboard('{Escape}')

    expect(firstInput()).toHaveValue('')
    expect(secondInput()).toHaveValue('')
    expect(result()).toHaveTextContent('0')
    expect(result()).not.toHaveAccessibleDescription()
    expect(within(screen.getByRole('region', { name: 'Recent' })).getAllByRole('listitem')).toHaveLength(1)
  })

  describe('keyboard shortcuts outside the form', () => {
    it('clears inputs and result on Escape with focus on the page body', async () => {
      respondWithResult(5)
      const user = setup()
      await fillOperands(user, '2', '3')
      await user.click(calculateButton())
      await waitFor(() => expect(result()).toHaveTextContent('5'))

      await user.click(document.body)
      expect(document.body).toHaveFocus()
      await user.keyboard('{Escape}')

      expect(firstInput()).toHaveValue('')
      expect(secondInput()).toHaveValue('')
      expect(result()).toHaveTextContent('0')
    })

    it('calculates on Enter with focus on the page body', async () => {
      respondWithResult(5)
      const user = setup()
      await fillOperands(user, '2', '3')

      await user.click(document.body)
      expect(document.body).toHaveFocus()
      await user.keyboard('{Enter}')

      await waitFor(() => expect(result()).toHaveTextContent('5'))
      expect(sentRequest()).toEqual({ url: '/api/v1/add', body: { a: 2, b: 3 } })
    })

    it('reuses a focused history item on Enter without calculating', async () => {
      respondWithResult(5)
      const user = setup()
      await fillOperands(user, '2', '3')
      await user.click(calculateButton())
      const item = await screen.findByRole('button', { name: '2 + 3 = 5, use as first number' })

      item.focus()
      await user.keyboard('{Enter}')

      expect(firstInput()).toHaveValue('5')
      expect(secondInput()).toHaveFocus()
      expect(fetchMock).toHaveBeenCalledTimes(1)
    })

    it('sends exactly one request on Enter inside an input', async () => {
      respondWithResult(5)
      const user = setup()
      await fillOperands(user, '2', '3')

      await user.keyboard('{Enter}')

      await waitFor(() => expect(result()).toHaveTextContent('5'))
      expect(fetchMock).toHaveBeenCalledTimes(1)
    })
  })

  it('clears the inputs with the Clear button', async () => {
    const user = setup()

    await fillOperands(user, '2', '3')
    await user.click(screen.getByRole('button', { name: 'Clear' }))

    expect(firstInput()).toHaveValue('')
    expect(secondInput()).toHaveValue('')
  })

  it('reuses a history result as the first operand, at full precision, and focuses the second input', async () => {
    respondWithResult(1 / 3)
    const user = setup()

    await fillOperands(user, '1', '3')
    await user.click(screen.getByRole('radio', { name: 'Divide' }))
    await user.click(calculateButton())
    await user.clear(firstInput())

    await user.click(await screen.findByRole('button', { name: '1 ÷ 3 = 0.333333333333, use as first number' }))

    expect(firstInput()).toHaveValue(String(1 / 3))
    expect(secondInput()).toHaveFocus()
  })

  it('focuses Calculate after reusing a history result while sqrt is selected', async () => {
    respondWithResult(9)
    const user = setup()

    await user.click(screen.getByRole('radio', { name: 'Square root' }))
    await fillOperands(user, '81')
    await user.click(calculateButton())
    // An empty operand disables Calculate, so it must be re-enabled before it can take focus.
    await user.clear(firstInput())
    expect(calculateButton()).toBeDisabled()

    await user.click(await screen.findByRole('button', { name: '√81 = 9, use as first number' }))

    expect(firstInput()).toHaveValue('9')
    expect(calculateButton()).toHaveFocus()
  })

  it('keeps only the last 5 calculations, newest first', async () => {
    fetchMock.mockImplementation((_, init) => {
      const { a } = JSON.parse(String(init?.body)) as { a: number }
      return Promise.resolve(jsonResponse(200, { operation: 'add', result: a + 1 }))
    })
    const user = setup()

    for (const a of ['1', '2', '3', '4', '5', '6']) {
      await fillOperands(user, a, '1')
      await user.click(calculateButton())
      await waitFor(() => expect(result()).toHaveAccessibleDescription(`${a} + 1`))
    }

    const items = within(screen.getByRole('region', { name: 'Recent' })).getAllByRole('button')
    expect(items.map((item) => item.getAttribute('aria-label'))).toEqual([
      '6 + 1 = 7, use as first number',
      '5 + 1 = 6, use as first number',
      '4 + 1 = 5, use as first number',
      '3 + 1 = 4, use as first number',
      '2 + 1 = 3, use as first number',
    ])
  })

  it.each<[number, string]>([
    [0.1 + 0.2, '0.3'],
    [2 / 3, '0.666666666667'],
    [1.23456789012345e21, '1.23456789012e+21'],
  ])('formats the API result %d as %s', async (value, expected) => {
    respondWithResult(value)
    const user = setup()

    await fillOperands(user, '1', '1')
    await user.click(calculateButton())

    await waitFor(() => expect(result()).toHaveTextContent(expected))
  })
})
