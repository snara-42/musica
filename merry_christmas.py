# pip install pygame music21

import music21
import pygame.midi

chor = """ tinynotation: 3/4 r4 r
c4 f f8 g f e d4 d8 r   d4 g g8 a g f e4 c8 r
c4 a a8 b- a g f4 d  c8 c d4 g e f2
c4 f f f e2    e4 f e d c2
g4 a g f c' c  c8 c d4 g e f2
c4 f8 r f g f e d r d r  d4 g8 r g a g f e r c r
c4 a8 r a b- a g f r d r  c c d r g r e r f2.
"""

bass = """ tinynotation: 3/4 r4 r
r F4 F F B- B- B-  G G G c c c
A4 A A d F A  B-4 G c  F  c8 B- A G
F4 A F c G c  B G B c C E
F4 C F A F A  B- G c F C8 D E C
F4 A F B- B-8 c B- A  G4 B G c c8 d c B-
A4 c# A d B-8 c B- A  B-4 G c F f2
"""

pygame.midi.init()
c = music21.converter.parse(chor)
b = music21.converter.parse(bass)
c.insert(music21.instrument.Vibraphone())
b.insert(music21.instrument.AcousticBass())

s = music21.stream.Score([c, b])
# s.show()
sp = music21.midi.realtime.StreamPlayer(s)
sp.play()
del sp
pygame.midi.quit()
exit()
