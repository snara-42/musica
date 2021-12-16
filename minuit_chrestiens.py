# pip install pygame music21

import music21
import pygame.midi

chor = """ tinynotation: 12/8  r1.
e4. e4 e8 g4.~ g4 g8 a4 a8 f4 a8 c'2. g4 g8 e4 d8 c4. e4 f8 g4. f4 d8 c1.
e4. e4 e8 g4.~ g4 g8 a4 a8 f4 a8 c'2. g4 g8 f#4 e8 b4. g4 a8 b4. c'4 b8 e2.~ e4.~ e4
g8 g4. a d g a4 g8 c'4 e8 a4. g4 g8 g4. a d g a4 g8 c'4 e8 g2.
c'2.~ c'4. b4 a8 b2.~ b4.~ b4 b8 d'2.~ d'4 a8 a4 a8 c'2. c'4.~ c'4 c'8
e'2. d'4.~ d'4 g8 c'2.~ c'4. b4 a8 g2.~ g4 g8 a4 g8 g2.~ g4.
c'4. d'2.~ d'4. g4. e'2.~ e'4. d'4. c'2. b4. c'4 d'8 c'1.
"""

piano = """ tinynotation: 12/8
E8 c G e c G E c G e c G
E c G e c G  E c G e c G  F c A f c A  E c G e c G E c G e c G
E c G e c G  F B G d B G  E c G e c G  E c G e c G
E c G e c G  E c G e c G  F c A f c A  E c G e c G E B- G e B- G
E B G e B G F# B A e- B A E B G e B G  E B G e B G
F B G d B G  F B G d B G  E c G e c G  E c G e c G
F B G d B G  F B G d B G  E c G e c G  E c G e c G
E c A e c A  E c A e c A  E B G e B G  E B G e B G
F d A f d A  F d A f d A  E c A e c A  E c A e c A
E c G e c G  D B G d B G  G e c g e c  F c A f c A
E c G e c G  F B G d B G  E c G e c G  E c G e c G
D B G d B G  D B G d B G  G e c g e c  F d A f d A
E c G e c G  F B G d B G  C E G c G E  c2.
"""

bass = """ tinynotation: 12/8
CC2.~ CC  CC CC FF CC  EE GG GGG
 CC~  CC  CC CC FF CC  CC BBB BBB EE~ EE
BB4. GG BB GG  C GG C GG
 BB  GG BB GG  C GG C~ C
AA2.~ AA EE~ EE DD~ DD AA~ AA
 GG  FF EE FF GG GG  CC4. EE GG C
G2. F E F G GG C1.
"""

pygame.midi.init()
c = music21.converter.parse(chor)
p = music21.converter.parse(piano)
b = music21.converter.parse(bass)
c.insert(music21.instrument.Oboe())
p.insert(music21.instrument.Lute())
b.insert(music21.instrument.StringInstrument())

s = music21.stream.Score([c, p, b])
# s.show()
sp = music21.midi.realtime.StreamPlayer(s)
sp.play()
del sp
pygame.midi.quit()
exit()
